package packet

import (
	c "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
)

type Enum interface {
	String() string
	Index() int
}

type ClientError byte

const (
	Ok ClientError = iota
	CallsignInUse
	CallsignInvalid
	Registered
	Syntax
	SourceCallsignInvalid
	AuthFail
	NoCallsignFound
	NoFlightPlan
	NoWeatherFound
	InvalidProtocolVision
	RequestLevelTooHigh
	ServerOverLoad
	UserBaned
)

var clientErrorsString = []string{"No error", "Callsign in use", "Invalid callsign", "Already registerd",
	"Syntax error", "Invalid source callsign", "Invalid CID/password", "No such callsign", "No flightplan",
	"No such weather profile", "Invalid protocol revision", "Requested level too high", "Too many clients connected",
	"CID/PID was suspended"}

func (e ClientError) String() string {
	return clientErrorsString[e]
}

func (e ClientError) Index() int {
	return int(e)
}

type Rating int

const (
	Ban Rating = iota - 1
	Normal
	Observer
	STU1
	STU2
	STU3
	CTR1
	CTR2
	CTR3
	Instructor1
	Instructor2
	Instructor3
	Supervisor
	Administrator
)

var ratingString = []string{"Baned", "Pilot", "Observer", "Ground/Delivery", "Tower Controller", "TMA Controller",
	"Enroute Controller", "Senior Controller", "Instructor 1", "Instructor 2", "Instructor 3",
	"Supervisor", "Administrator"}

func (r Rating) String() string {
	return ratingString[r]
}

func (r Rating) Index() int {
	return int(r)
}

type Facility byte

const (
	Pilot Facility = 1 << iota
	OBS
	DEL
	GND
	TWR
	APP
	CTR
	FSS
)

var facilityString = []string{"Pilot", "OBS", "DEL", "GND", "TWR", "APP", "FSS"}

func (f Facility) String() string {
	return facilityString[f]
}

func (f Facility) Index() int {
	return int(f)
}

func (f Facility) CheckFacility(facility Facility) bool {
	return f&facility != 0
}

var ratingFacilityMap = map[Rating]Facility{
	Ban:           0,
	Normal:        Pilot,
	Observer:      Pilot | OBS,
	STU1:          Pilot | OBS | DEL | GND,
	STU2:          Pilot | OBS | DEL | GND | TWR,
	STU3:          Pilot | OBS | DEL | GND | TWR | APP,
	CTR1:          Pilot | OBS | DEL | GND | TWR | APP | CTR,
	CTR2:          Pilot | OBS | DEL | GND | TWR | APP | CTR,
	CTR3:          Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
	Instructor1:   Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
	Instructor2:   Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
	Instructor3:   Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
	Supervisor:    Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
	Administrator: Pilot | OBS | DEL | GND | TWR | APP | CTR | FSS,
}

func (r Rating) CheckRatingFacility(facility Facility) bool {
	return ratingFacilityMap[r].CheckFacility(facility)
}

func SyncRatingConfig() {
	config, _ := c.GetConfig()
	if len(config.RatingConfig) == 0 {
		return
	}
	for rating, facility := range config.RatingConfig {
		ratingFacilityMap[Rating(utils.StrToInt(rating, 0))] = Facility(facility)
	}
}

type ClientCommand string

var (
	AddAtc          = ClientCommand("#AA")
	RemoveAtc       = ClientCommand("#DA")
	AddPilot        = ClientCommand("#AP")
	RemovePilot     = ClientCommand("#DP")
	RequestHandoff  = ClientCommand("#HO")
	AcceptHandoff   = ClientCommand("#HA")
	ProController   = ClientCommand("#PC")
	PilotPosition   = ClientCommand("@")
	AtcPosition     = ClientCommand("%")
	AtcSubVisPoint  = ClientCommand("'")
	Message         = ClientCommand("#TM")
	WindDelta       = ClientCommand("#DL")
	Ping            = ClientCommand("$PI")
	Pong            = ClientCommand("$PO")
	Plan            = ClientCommand("$FP")
	AtcEditPlan     = ClientCommand("$AM")
	KillClient      = ClientCommand("$!!")
	Error           = ClientCommand("$ER")
	ClientQuery     = ClientCommand("$CQ")
	ClientResponse  = ClientCommand("$CR")
	SquawkBox       = ClientCommand("$SB")
	RequestWeather  = ClientCommand("$RW")
	ResponseWeather = ClientCommand("$WX")
	CloudData       = ClientCommand("$CD")
	WindData        = ClientCommand("$WD")
	TempData        = ClientCommand("$TD")
	RequestComm     = ClientCommand("$C?")
	ReplyComm       = ClientCommand("$CI")
	RequestAcars    = ClientCommand("$AX")
	ReplyAcars      = ClientCommand("$AR")
)

var PossibleClientCommands = [][]byte{[]byte(PilotPosition), []byte(AtcPosition), []byte(AtcSubVisPoint),
	[]byte(Message), []byte(ClientQuery), []byte(ClientResponse), []byte(RequestHandoff), []byte(AcceptHandoff),
	[]byte(ProController), []byte(AddAtc), []byte(RemoveAtc), []byte(AddPilot), []byte(RemovePilot)}

func (c ClientCommand) String() string {
	return string(c)
}

func (c ClientCommand) Index() int {
	return 0
}

type BroadcastTarget string

var (
	AllClient BroadcastTarget = "*"
	AllATC    BroadcastTarget = "*A"
	AllPilot  BroadcastTarget = "*P"
)

func (b BroadcastTarget) String() string {
	return string(b)
}

func (b BroadcastTarget) Index() int {
	return 0
}
