package packet

import (
	"errors"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/database"
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
	IsAtc       bool
	Callsign    string
	Rating      Rating
	Facility    Facility
	Cid         database.UserId
	Protocol    int
	RealName    string
	Socket      *ConnectionHandler
	Position    [4]Position
	SimType     int
	Transponder int
	Altitude    int
	GroundSpeed int
	Frequency   int
	VisualRange float64
	FlightPlan  *database.FlightPlan
	AtisInfo    []string
	OnlineTime  time.Time
}

func NewClient(callsign string, rating Rating, cid database.UserId, protocol int, realName string, socket *ConnectionHandler, isAtc bool) *Client {
	flightPlan, _ := database.GetFlightPlan(callsign)
	return &Client{
		IsAtc:       isAtc,
		Callsign:    callsign,
		Rating:      rating,
		Facility:    0,
		Cid:         cid,
		Protocol:    protocol,
		RealName:    realName,
		Socket:      socket,
		Position:    [4]Position{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		SimType:     0,
		Transponder: 9999,
		Altitude:    0,
		GroundSpeed: 0,
		Frequency:   99998,
		VisualRange: 40,
		FlightPlan:  flightPlan,
		AtisInfo:    make([]string, 0, 4),
		OnlineTime:  time.Now(),
	}
}

func (c *Client) UpdateFlightPlan(flightPlanData []string) error {
	if c.FlightPlan == nil {
		flightPlan, err := database.CreateFlightPlan(c.Callsign, flightPlanData)
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
	err := c.FlightPlan.UpdateFlightPlan(flightPlanData)
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
