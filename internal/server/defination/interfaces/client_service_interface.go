// Package interfaces
package interfaces

import (
	. "github.com/half-nothing/fsd-server/internal/server/defination/fsd"
)

type ClientServiceInterface interface {
	GetOnlineClient() *ApiResponse[ResponseOnlineClient]
	SendMessageToClient(req *RequestSendMessageToClient) *ApiResponse[ResponseSendMessageToClient]
	KillClient(req *RequestKillClient) *ApiResponse[ResponseKillClient]
}

type OnlineGeneral struct {
	Version          int    `json:"version"`
	GenerateTime     string `json:"generate_time"`
	ConnectedClients int    `json:"connected_clients"`
	OnlinePilot      int    `json:"online_pilot"`
	OnlineController int    `json:"online_controller"`
}

type OnlinePilot struct {
	Cid         int     `json:"cid"`
	Callsign    string  `json:"callsign"`
	RealName    string  `json:"real_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Transponder string  `json:"transponder"`
	Altitude    int     `json:"altitude"`
	GroundSpeed int     `json:"ground_speed"`
	LogonTime   string  `json:"logon_time"`
}

type OnlineController struct {
	Cid       int      `json:"cid"`
	Callsign  string   `json:"callsign"`
	RealName  string   `json:"real_name"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Rating    int      `json:"rating"`
	Facility  int      `json:"facility"`
	Frequency int      `json:"frequency"`
	AtcInfo   []string `json:"atc_info"`
	LogonTime string   `json:"logon_time"`
}
type OnlineClients struct {
	General     OnlineGeneral       `json:"general"`
	Pilots      []*OnlinePilot      `json:"pilots"`
	Controllers []*OnlineController `json:"controllers"`
	Facilities  *[]FacilityModel    `json:"facilities"`
	Ratings     *[]RatingModel      `json:"ratings"`
}

type ResponseOnlineClient struct {
	OnlineClients
}

type RequestSendMessageToClient struct {
	JwtHeader
	Cid     int
	SendTo  string `param:"callsign"`
	Message string `json:"message"`
}

type ResponseSendMessageToClient bool

type RequestKillClient struct {
	JwtHeader
	TargetCallsign string `param:"callsign"`
	Reason         string `json:"reason"`
}

type ResponseKillClient bool
