// Package interfaces
package interfaces

import (
	"github.com/half-nothing/fsd-server/internal/server/database"
)

type ActivityModel struct {
	Id         uint   `json:"id"`
	Publisher  int    `json:"publisher"`
	Title      string `json:"title"`
	ImageUrl   string `json:"image_url"`
	ActiveTime string `json:"active_time"`
	Departure  string `json:"departure"`
	Arrival    string `json:"arrival"`
	Route      string `json:"route"`
	Distance   int    `json:"distance"`
	Status     int    `json:"status"`
	NOTAMS     string `json:"notams"`
}

type ActivityServiceInterface interface {
	GetActivities(req *RequestGetActivities) *ApiResponse[ResponseGetActivities]
	GetActivityInfo(req *RequestActivityInfo) *ApiResponse[ResponseActivityInfo]
	AddActivity(req *RequestAddActivity) *ApiResponse[ResponseAddActivity]
	DeleteActivity(req *RequestDeleteActivity) *ApiResponse[ResponseDeleteActivity]
	EditActivity(req *RequestEditActivity) *ApiResponse[ResponseEditActivity]
	ControllerJoin(req *RequestControllerJoin) *ApiResponse[ResponseControllerJoin]
	ControllerLeave(req *RequestControllerLeave) *ApiResponse[ResponseControllerLeave]
	PilotJoin(req *RequestPilotJoin) *ApiResponse[ResponsePilotJoin]
	PilotLeave(req *RequestPilotLeave) *ApiResponse[ResponsePilotLeave]
	EditPilotStatus(req *RequestEditPilotStatus) *ApiResponse[ResponseEditPilotStatus]
}

type RequestGetActivities struct {
	Time string `query:"time"`
}

type ResponseGetActivities struct {
	Items []*database.Activity `json:"items"`
}

type RequestActivityInfo struct {
	ActivityId uint `param:"id"`
}

type ResponseActivityInfo database.Activity

type RequestAddActivity struct {
	JwtHeader
	Cid int
	*database.Activity
}

type ResponseAddActivity struct {
	*database.Activity
}

type RequestDeleteActivity struct{}
type ResponseDeleteActivity struct{}
type RequestEditActivity struct{}
type ResponseEditActivity struct{}
type RequestControllerJoin struct{}
type ResponseControllerJoin struct{}
type RequestControllerLeave struct{}
type ResponseControllerLeave struct{}
type RequestPilotJoin struct{}
type ResponsePilotJoin struct{}
type RequestPilotLeave struct{}
type ResponsePilotLeave struct{}
type RequestEditPilotStatus struct{}
type ResponseEditPilotStatus struct{}
