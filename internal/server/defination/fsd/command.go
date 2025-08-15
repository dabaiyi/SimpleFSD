// Package fsd
package fsd

type ClientCommand string

var (
	AddAtc          = ClientCommand("#AA")
	RemoveAtc       = ClientCommand("#DA")
	AddPilot        = ClientCommand("#AP")
	RemovePilot     = ClientCommand("#DP")
	ProController   = ClientCommand("#PC")
	PilotPosition   = ClientCommand("@")
	AtcPosition     = ClientCommand("%")
	AtcSubVisPoint  = ClientCommand("'")
	Message         = ClientCommand("#TM")
	WindDelta       = ClientCommand("#DL")
	RequestHandoff  = ClientCommand("$HO")
	AcceptHandoff   = ClientCommand("$HA")
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

type CommandRequirement struct {
	RequireLength int
	Fatal         bool
}

func (c ClientCommand) String() string {
	return string(c)
}

func (c ClientCommand) Index() int {
	return 0
}
