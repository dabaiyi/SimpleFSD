package packet

import (
	"bytes"
	"errors"
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	. "github.com/half-nothing/fsd-server/internal/server/defination/fsd"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	isAtc          bool
	callsign       string
	rating         Rating
	facility       Facility
	user           *database.User
	protocol       int
	realName       string
	socket         ConnectionHandlerInterface
	position       [4]Position
	simType        int
	transponder    string
	altitude       int
	groundSpeed    int
	frequency      int
	visualRange    float64
	flightPlan     *database.FlightPlan
	atisInfo       []string
	history        *database.History
	disconnect     atomic.Bool
	motdBytes      []byte
	reconnectTimer *time.Timer
	lock           sync.RWMutex
}

func NewClient(callsign string, rating Rating, protocol int, realName string, socket ConnectionHandlerInterface, isAtc bool) *Client {
	socket.SetCallsign(callsign)
	var flightPlan *database.FlightPlan = nil
	if !isAtc && !config.Server.General.SimulatorServer {
		var err error
		flightPlan, err = database.GetFlightPlan(socket.User().Cid)
		if errors.Is(err, database.ErrFlightPlanNotFound) {
			logger.WarnF("No flight plan found for %s(%d)", callsign, socket.User().Cid)
		} else if err != nil {
			logger.WarnF("Fail to get flight plan for %s(%d): %v", callsign, socket.User().Cid, err)
		}
	}
	return &Client{
		isAtc:          isAtc,
		callsign:       callsign,
		rating:         rating,
		facility:       0,
		user:           socket.User(),
		protocol:       protocol,
		realName:       realName,
		socket:         socket,
		position:       [4]Position{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		simType:        0,
		transponder:    "2000",
		altitude:       0,
		groundSpeed:    0,
		frequency:      99998,
		visualRange:    40,
		flightPlan:     flightPlan,
		atisInfo:       make([]string, 0, 4),
		history:        database.NewHistory(socket.User().Cid, callsign, isAtc),
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
		logger.InfoF("[%s](%s) client session deleted", c.socket.ConnId(), c.callsign)

		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
			c.reconnectTimer = nil
		}

		if c.isAtc || !config.Server.General.SimulatorServer {
			if err := c.history.End(); err != nil {
				logger.ErrorF("[%s](%s) Failed to end history: %v", c.socket.ConnId(), c.callsign, err)
			}
		}

		if c.isAtc {
			if err := c.user.AddAtcTime(c.history.OnlineTime); err != nil {
				logger.ErrorF("[%s](%s) Failed to add ATC time: %v", c.socket.ConnId(), c.callsign, err)
			}
		} else if !config.Server.General.SimulatorServer {
			// 如果不是模拟机服务器, 则写入机组连线时长
			if err := c.user.AddPilotTime(c.history.OnlineTime); err != nil {
				logger.ErrorF("[%s](%s) Failed to add pilot time: %v", c.socket.ConnId(), c.callsign, err)
			}
		}

		if !clientManager.DeleteClient(c.callsign) {
			logger.ErrorF("[%s](%s) Failed to delete from client manager", c.socket.ConnId(), c.callsign)
		}
	}
}

func (c *Client) Reconnect(socket ConnectionHandlerInterface) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.disconnect.Load() {
		return false
	}

	logger.InfoF("[%s](%s) client reconnected", c.socket.ConnId, c.callsign)

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}

	c.ClearAtcAtisInfo()
	c.disconnect.Store(false)
	c.socket = socket
	socket.SetCallsign(c.callsign)
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
	if c.socket.Conn() != nil {
		_ = c.socket.Conn().Close()
	}

	// 取消之前的定时器
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	if immediate {
		return
	}

	c.reconnectTimer = time.AfterFunc(config.Server.FSDServer.SessionCleanDuration, c.Delete)
	logger.InfoF("[%s](%s) client disconnected, reconnect window: %v", c.socket.ConnId,
		c.callsign, config.Server.FSDServer.SessionCleanDuration)
}

func (c *Client) UpsertFlightPlan(flightPlanData []string) error {
	if c.flightPlan == nil {
		flightPlan, err := database.UpsertFlightPlan(c.user, c.callsign, flightPlanData)
		if err != nil {
			return err
		}
		c.flightPlan = flightPlan
		return nil
	}
	// 如果是模拟机服务器, 只创建就行
	if config.Server.General.SimulatorServer {
		return nil
	}
	if c.flightPlan.Locked {
		departureAirport := flightPlanData[5]
		arrivalAirport := flightPlanData[9]
		if c.flightPlan.DepartureAirport != departureAirport || c.flightPlan.ArrivalAirport != arrivalAirport {
			c.flightPlan.Locked = false
		}
	}
	err := c.flightPlan.UpdateFlightPlan(flightPlanData, false)
	return err
}

func (c *Client) SetPosition(index int, lat float64, lon float64) error {
	if index >= 4 {
		return errors.New("position index out of range")
	}
	c.position[index].Latitude = lat
	c.position[index].Longitude = lon
	return nil
}

func (c *Client) UpdatePilotPos(transponder int, lat float64, lon float64, alt int, groundSpeed int) {
	_ = c.SetPosition(0, lat, lon)
	c.transponder = fmt.Sprintf("%04d", transponder)
	c.altitude = alt
	c.groundSpeed = groundSpeed
}

func (c *Client) UpdateAtcPos(frequency int, facility Facility, visualRange float64, lat float64, lon float64) {
	_ = c.SetPosition(0, lat, lon)
	c.frequency = frequency
	c.facility = facility
	c.visualRange = visualRange
}

func (c *Client) UpdateAtcVisPoint(visIndex int, lat float64, lon float64) error {
	if visIndex < 0 || visIndex > 2 {
		return errors.New("visIndex out of range [0,2]")
	}
	return c.SetPosition(visIndex+1, lat, lon)
}

func (c *Client) ClearAtcAtisInfo() {
	c.atisInfo = c.atisInfo[:0]
}

func (c *Client) AddAtcAtisInfo(atisInfo string) {
	c.atisInfo = append(c.atisInfo, atisInfo)
}

func (c *Client) SendError(result *Result) {
	if result.Success {
		return
	}

	packet := makePacket(Error, "server", c.callsign, fmt.Sprintf("%03d", result.Errno.Index()), result.Env, result.Errno.String())
	c.SendLine(packet)

	if result.Fatal {
		c.socket.SetDisconnected(true)
		c.disconnect.Store(true)
		time.AfterFunc(500*time.Millisecond, func() {
			if !clientManager.DeleteClient(c.callsign) {
				logger.ErrorF("[%s](%s) Failed to delete from client manager", c.socket.ConnId, c.callsign)
			}
		})
	}
}

func (c *Client) SendLineWithoutLog(line []byte) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.disconnect.Load() {
		logger.WarnF("[%s](%s) Attempted send to disconnected client", c.socket.ConnId, c.callsign)
		return
	}

	if !bytes.HasSuffix(line, splitSign) {
		line = append(line, splitSign...)
	}

	if _, err := c.socket.Conn().Write(line); err != nil {
		logger.ErrorF("[%s](%s) Failed to send data: %v", c.socket.ConnId, c.callsign, err)
	}
}

func (c *Client) SendLine(line []byte) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.disconnect.Load() {
		logger.DebugF("[%s](%s) Attempted send to disconnected client", c.socket.ConnId, c.callsign)
		return
	}

	if !bytes.HasSuffix(line, splitSign) {
		logger.DebugF("[%s](%s) <- %s", c.socket.ConnId, c.callsign, line)
		line = append(line, splitSign...)
	} else {
		logger.DebugF("[%s](%s) <- %s", c.socket.ConnId, c.callsign, line[:len(line)-splitSignLen])
	}

	if _, err := c.socket.Conn().Write(line); err != nil {
		logger.WarnF("[%s](%s) Failed to send data: %v", c.socket.ConnId, c.callsign, err)
	}
}

func (c *Client) SendMotd() {
	if c.motdBytes != nil {
		c.SendLine(c.motdBytes)
		return
	}

	data := make([][]byte, 0, len(config.Server.FSDServer.Motd)+1)
	data = append(data, []byte(fmt.Sprintf("%sserver:%s:Welcome to use %s v%s\r\n",
		Message, c.callsign, config.Server.FSDServer.FSDName, logger.AppVersion.String())))

	for _, message := range config.Server.FSDServer.Motd {
		data = append(data, makePacket(Message, "server", c.callsign, message))
	}

	buffer := bytes.Buffer{}
	for _, msg := range data {
		buffer.Write(msg)
	}
	c.motdBytes = buffer.Bytes()
	c.SendLine(c.motdBytes)
}

func (c *Client) CheckFacility(facility Facility) bool {
	return facility.CheckFacility(c.facility)
}

func (c *Client) CheckRating(rating []Rating) bool {
	return slices.Contains(rating, c.rating)
}

func (c *Client) IsAtc() bool { return c.isAtc }

func (c *Client) Callsign() string { return c.callsign }

func (c *Client) Rating() Rating { return c.rating }

func (c *Client) Facility() Facility { return c.facility }

func (c *Client) RealName() string { return c.realName }

func (c *Client) Position() [4]Position { return c.position }

func (c *Client) VisualRange() float64 { return c.visualRange }

func (c *Client) SetUser(user *database.User) { c.user = user }

func (c *Client) SetSimType(simType int) { c.simType = simType }

func (c *Client) FlightPlan() *database.FlightPlan { return c.flightPlan }

func (c *Client) User() *database.User { return c.user }

func (c *Client) Frequency() int { return c.frequency }

func (c *Client) AtisInfo() []string { return c.atisInfo }

func (c *Client) History() *database.History { return c.history }

func (c *Client) Transponder() string { return c.transponder }

func (c *Client) Altitude() int { return c.altitude }

func (c *Client) GroundSpeed() int { return c.groundSpeed }
