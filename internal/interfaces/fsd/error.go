// Package fsd
package fsd

type ClientError byte

const (
	CommandOk ClientError = iota
	CallsignInUse
	CallsignInvalid
	Syntax
	SourceCallsignInvalid
	AuthFail
	NoCallsignFound
	NoFlightPlan
	InvalidProtocolVision
	RequestLevelTooHigh
	UserBaned
)

var clientErrorsString = []string{"No error", "callsign in use", "Invalid callsign",
	"Syntax error", "Invalid source callsign", "Invalid CID/password", "No such callsign", "No flightplan",
	"Invalid protocol revision", "Requested level too high", "CID/PID was suspended"}

func (e ClientError) String() string {
	return clientErrorsString[e]
}

func (e ClientError) Index() int {
	return int(e)
}
