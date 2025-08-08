package packet

import (
	"bytes"
	"errors"
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"net"
	"sync"
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
	Socket         net.Conn
	Position       [4]Position
	SimType        int
	Transponder    int
	Altitude       int
	GroundSpeed    int
	Frequency      int
	VisualRange    float64
	FlightPlan     *database.FlightPlan
	AtisInfo       []string
	OnlineTime     time.Time
	disconnect     bool
	motdBytes      []byte
	reconnectTimer *time.Timer
	lock           sync.Mutex
}

func NewClient(callsign string, rating Rating, user *database.User, protocol int, realName string, socket net.Conn, isAtc bool) *Client {
	var flightPlan *database.FlightPlan = nil
	if !isAtc {
		flightPlan, _ = database.GetFlightPlan(user.Cid)
	} else {
		flightPlan = nil
	}
	return &Client{
		IsAtc:          isAtc,
		Callsign:       callsign,
		Rating:         rating,
		Facility:       0,
		User:           user,
		Protocol:       protocol,
		RealName:       realName,
		Socket:         socket,
		Position:       [4]Position{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		SimType:        0,
		Transponder:    9999,
		Altitude:       0,
		GroundSpeed:    0,
		Frequency:      99998,
		VisualRange:    40,
		FlightPlan:     flightPlan,
		AtisInfo:       make([]string, 0, 4),
		OnlineTime:     time.Now(),
		motdBytes:      nil,
		disconnect:     false,
		reconnectTimer: nil,
		lock:           sync.Mutex{},
	}
}

func (c *Client) Disconnected() bool {
	return c.disconnect
}

func (c *Client) Delete() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.disconnect {
		logger.DebugF("[%s] Client session deleted", c.Callsign)
		_ = clientManager.DeleteClient(c.Callsign)
	}
}

func (c *Client) Reconnect(socket net.Conn) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.disconnect {
		return false
	}

	logger.DebugF("[%s] Client reconnected", c.Callsign)

	c.reconnectTimer.Stop()
	c.disconnect = false
	c.Socket = socket
	c.reconnectTimer = nil
	return true
}

func (c *Client) MarkedDisconnect() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.disconnect {
		return
	}

	c.disconnect = true

	if c.Socket != nil {
		c.Socket = nil
	}

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	c.reconnectTimer = time.AfterFunc(config.SessionCleanDuration, c.Delete)
}

func (c *Client) UpdateFlightPlan(flightPlanData []string) error {
	if c.FlightPlan == nil {
		flightPlan, err := database.CreateFlightPlan(c.User, c.Callsign, flightPlanData)
		if err != nil {
			return err
		}
		c.FlightPlan = flightPlan
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
		_ = c.Socket.Close()
	}
}

func (c *Client) SendLine(line []byte) {
	if !bytes.HasSuffix(line, splitSign) {
		logger.DebugF("[%s] <- %s", c.Callsign, line)
		line = append(line, splitSign...)
	} else {
		logger.DebugF("[%s] <- %s", c.Callsign, line[:len(line)-splitSignLen])
	}
	if c.Socket != nil {
		_, _ = c.Socket.Write(line)
	}
}

func (c *Client) SendMotd() {
	if c.motdBytes != nil {
		_, _ = c.Socket.Write(c.motdBytes)
		return
	}
	data := make([][]byte, 0, len(config.ServerConfig.Motd)+1)
	data = append(data, []byte(fmt.Sprintf("%sserver:%s:Welcome to user %s v%s\r\n", Message, c.Callsign, config.AppName, config.AppVersion)))
	for _, message := range config.ServerConfig.Motd {
		data = append(data, makePacket(Message, "server", c.Callsign, message))
	}
	buffer := bytes.Buffer{}
	for _, msg := range data {
		buffer.Write(msg)
	}
	c.motdBytes = buffer.Bytes()
	c.SendLine(c.motdBytes)
}
