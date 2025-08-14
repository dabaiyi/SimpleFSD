package packet

import (
	"bytes"
	"errors"
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Position struct {
	Latitude  float64
	Longitude float64
}

func (p *Position) PositionValid() bool {
	return p.Latitude != 0 && p.Longitude != 0
}

type Client struct {
	IsAtc          bool
	Callsign       string
	Rating         Rating
	Facility       Facility
	User           *database.User
	Protocol       int
	RealName       string
	Socket         *ConnectionHandler
	Position       [4]Position
	SimType        int
	Transponder    int
	Altitude       int
	GroundSpeed    int
	Frequency      int
	VisualRange    float64
	FlightPlan     *database.FlightPlan
	AtisInfo       []string
	History        *database.History
	disconnect     atomic.Bool
	motdBytes      []byte
	reconnectTimer *time.Timer
	lock           sync.RWMutex
}

func NewClient(callsign string, rating Rating, protocol int, realName string, socket *ConnectionHandler, isAtc bool) *Client {
	socket.Callsign = callsign
	var flightPlan *database.FlightPlan = nil
	if !isAtc && !config.Server.General.SimulatorServer {
		var err error
		flightPlan, err = database.GetFlightPlan(socket.User.Cid)
		if errors.Is(err, database.ErrFlightPlanNotFound) {
			logger.WarnF("No flight plan found for %s(%d)", callsign, socket.User.Cid)
		} else if err != nil {
			logger.WarnF("Fail to get flight plan for %s(%d): %v", callsign, socket.User.Cid, err)
		}
	}
	return &Client{
		IsAtc:          isAtc,
		Callsign:       callsign,
		Rating:         rating,
		Facility:       0,
		User:           socket.User,
		Protocol:       protocol,
		RealName:       realName,
		Socket:         socket,
		Position:       [4]Position{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		SimType:        0,
		Transponder:    2000,
		Altitude:       0,
		GroundSpeed:    0,
		Frequency:      99998,
		VisualRange:    40,
		FlightPlan:     flightPlan,
		AtisInfo:       make([]string, 0, 4),
		History:        database.NewHistory(socket.User.Cid, callsign, isAtc),
		motdBytes:      nil,
		disconnect:     atomic.Bool{},
		reconnectTimer: nil,
		lock:           sync.RWMutex{},
	}
}

func (c *Client) Disconnected() bool {
	return c.disconnect.Load()
}

func (c *Client) Delete() {
	if c.disconnect.Load() {
		c.lock.Lock()
		defer c.lock.Unlock()
		logger.InfoF("[%s](%s) Client session deleted", c.Socket.ConnId, c.Callsign)

		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
			c.reconnectTimer = nil
		}

		if c.IsAtc || !config.Server.General.SimulatorServer {
			if err := c.History.End(); err != nil {
				logger.ErrorF("[%s](%s) Failed to end history: %v", c.Socket.ConnId, c.Callsign, err)
			}
		}

		if c.IsAtc {
			if err := c.User.AddAtcTime(c.History.OnlineTime); err != nil {
				logger.ErrorF("[%s](%s) Failed to add ATC time: %v", c.Socket.ConnId, c.Callsign, err)
			}
		} else if !config.Server.General.SimulatorServer {
			// 如果不是模拟机服务器, 则写入机组连线时长
			if err := c.User.AddPilotTime(c.History.OnlineTime); err != nil {
				logger.ErrorF("[%s](%s) Failed to add pilot time: %v", c.Socket.ConnId, c.Callsign, err)
			}
		}

		if !clientManager.DeleteClient(c.Callsign) {
			logger.ErrorF("[%s](%s) Failed to delete from client manager", c.Socket.ConnId, c.Callsign)
		}
	}
}

func (c *Client) Reconnect(socket *ConnectionHandler) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.disconnect.Load() {
		return false
	}

	logger.InfoF("[%s](%s) Client reconnected", c.Socket.ConnId, c.Callsign)

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}

	c.ClearAtcAtisInfo()
	c.disconnect.Store(false)
	c.Socket = socket
	socket.Callsign = c.Callsign
	return true
}

func (c *Client) MarkedDisconnect(immediate bool) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
		if immediate {
			c.Delete()
		}
	}()

	if !c.disconnect.CompareAndSwap(false, true) {
		return
	}

	// 关闭连接
	if c.Socket.Conn != nil {
		_ = c.Socket.Conn.Close()
	}

	// 取消之前的定时器
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	if immediate {
		return
	}

	c.reconnectTimer = time.AfterFunc(config.Server.FSDServer.SessionCleanDuration, c.Delete)
	logger.InfoF("[%s](%s) Client disconnected, reconnect window: %v", c.Socket.ConnId,
		c.Callsign, config.Server.FSDServer.SessionCleanDuration)
}

func (c *Client) UpsertFlightPlan(flightPlanData []string) error {
	if c.FlightPlan == nil {
		flightPlan, err := database.UpsertFlightPlan(c.User, c.Callsign, flightPlanData)
		if err != nil {
			return err
		}
		c.FlightPlan = flightPlan
		return nil
	}
	// 如果是模拟机服务器, 只创建就行
	if config.Server.General.SimulatorServer {
		return nil
	}
	if c.FlightPlan.Locked {
		departureAirport := flightPlanData[5]
		arrivalAirport := flightPlanData[9]
		if c.FlightPlan.DepartureAirport != departureAirport || c.FlightPlan.ArrivalAirport != arrivalAirport {
			c.FlightPlan.Locked = false
		}
	}
	err := c.FlightPlan.UpdateFlightPlan(flightPlanData, false)
	return err
}

func (c *Client) SetPosition(index int, lat float64, lon float64) error {
	if index >= 4 {
		return errors.New("position index out of range")
	}
	c.Position[index].Latitude = lat
	c.Position[index].Longitude = lon
	return nil
}

func (c *Client) UpdatePilotPos(transponder int, lat float64, lon float64, alt int, groundSpeed int) {
	_ = c.SetPosition(0, lat, lon)
	c.Transponder = transponder
	c.Altitude = alt
	c.GroundSpeed = groundSpeed
}

func (c *Client) UpdateAtcPos(frequency int, facility Facility, visualRange float64, lat float64, lon float64) {
	_ = c.SetPosition(0, lat, lon)
	c.Frequency = frequency
	c.Facility = facility
	c.VisualRange = visualRange
}

func (c *Client) UpdateAtcVisPoint(visIndex int, lat float64, lon float64) error {
	if visIndex < 0 || visIndex > 2 {
		return errors.New("visIndex out of range [0,2]")
	}
	return c.SetPosition(visIndex+1, lat, lon)
}

func (c *Client) ClearAtcAtisInfo() {
	c.AtisInfo = c.AtisInfo[:0]
}

func (c *Client) AddAtcAtisInfo(atisInfo string) {
	c.AtisInfo = append(c.AtisInfo, atisInfo)
}

func (c *Client) SendError(result *Result) {
	if result.success {
		return
	}

	packet := makePacket(Error, "server", c.Callsign, fmt.Sprintf("%03d", result.errno.Index()), result.env, result.errno.String())
	c.SendLine(packet)

	if result.fatal {
		c.Socket.Disconnected.Store(true)
		c.disconnect.Store(true)
		time.AfterFunc(500*time.Millisecond, func() {
			if !clientManager.DeleteClient(c.Callsign) {
				logger.ErrorF("[%s](%s) Failed to delete from client manager", c.Socket.ConnId, c.Callsign)
			}
		})
	}
}

func (c *Client) SendLineWithoutLog(line []byte) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.disconnect.Load() {
		logger.WarnF("[%s](%s) Attempted send to disconnected client", c.Socket.ConnId, c.Callsign)
		return
	}

	if !bytes.HasSuffix(line, splitSign) {
		line = append(line, splitSign...)
	}

	if _, err := c.Socket.Conn.Write(line); err != nil {
		logger.ErrorF("[%s](%s) Failed to send data: %v", c.Socket.ConnId, c.Callsign, err)
	}
}

func (c *Client) SendLine(line []byte) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.disconnect.Load() {
		logger.DebugF("[%s](%s) Attempted send to disconnected client", c.Socket.ConnId, c.Callsign)
		return
	}

	if !bytes.HasSuffix(line, splitSign) {
		logger.DebugF("[%s](%s) <- %s", c.Socket.ConnId, c.Callsign, line)
		line = append(line, splitSign...)
	} else {
		logger.DebugF("[%s](%s) <- %s", c.Socket.ConnId, c.Callsign, line[:len(line)-splitSignLen])
	}

	if _, err := c.Socket.Conn.Write(line); err != nil {
		logger.WarnF("[%s](%s) Failed to send data: %v", c.Socket.ConnId, c.Callsign, err)
	}
}

func (c *Client) SendMotd() {
	if c.motdBytes != nil {
		c.SendLine(c.motdBytes)
		return
	}

	data := make([][]byte, 0, len(config.Server.FSDServer.Motd)+1)
	data = append(data, []byte(fmt.Sprintf("%sserver:%s:Welcome to use %s v%s\r\n",
		Message, c.Callsign, config.Server.FSDServer.FSDName, logger.AppVersion.String())))

	for _, message := range config.Server.FSDServer.Motd {
		data = append(data, makePacket(Message, "server", c.Callsign, message))
	}

	buffer := bytes.Buffer{}
	for _, msg := range data {
		buffer.Write(msg)
	}
	c.motdBytes = buffer.Bytes()
	c.SendLine(c.motdBytes)
}

func (c *Client) CheckFacility(facility Facility) bool {
	return facility.CheckFacility(c.Facility)
}

func (c *Client) CheckRating(rating []Rating) bool {
	return slices.Contains(rating, c.Rating)
}
