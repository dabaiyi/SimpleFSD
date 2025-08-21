// Package fsd
package fsd

import (
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
)

type ClientInterface interface {
	Disconnected() bool
	Delete()
	Reconnect(socket ConnectionHandlerInterface) bool
	MarkedDisconnect(immediate bool)
	UpsertFlightPlan(flightPlanData []string) error
	SetPosition(index int, lat float64, lon float64) error
	UpdatePilotPos(transponder int, lat float64, lon float64, alt int, groundSpeed int, pbh uint32)
	UpdateAtcPos(frequency int, facility Facility, visualRange float64, lat float64, lon float64)
	UpdateAtcVisPoint(visIndex int, lat float64, lon float64) error
	ClearAtcAtisInfo()
	AddAtcAtisInfo(atisInfo string)
	SendError(result *Result)
	SendLineWithoutLog(line []byte)
	SendLine(line []byte)
	SendMotd()
	CheckFacility(facility Facility) bool
	CheckRating(rating []Rating) bool
	IsAtc() bool
	Callsign() string
	Rating() Rating
	Facility() Facility
	RealName() string
	Position() [4]Position
	VisualRange() float64
	SetUser(user *operation.User)
	SetSimType(simType int)
	FlightPlan() *operation.FlightPlan
	User() *operation.User
	Frequency() int
	AtisInfo() []string
	History() *operation.History
	Transponder() string
	Altitude() int
	GroundSpeed() int
	Heading() int
}
