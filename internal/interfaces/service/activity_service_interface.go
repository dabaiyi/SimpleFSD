// Package service
package service

import (
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	"time"
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
	GetActivitiesPage(req *RequestGetActivitiesPage) *ApiResponse[ResponseGetActivitiesPage]
	GetActivityInfo(req *RequestActivityInfo) *ApiResponse[ResponseActivityInfo]
	AddActivity(req *RequestAddActivity) *ApiResponse[ResponseAddActivity]
	DeleteActivity(req *RequestDeleteActivity) *ApiResponse[ResponseDeleteActivity]
	ControllerJoin(req *RequestControllerJoin) *ApiResponse[ResponseControllerJoin]
	ControllerLeave(req *RequestControllerLeave) *ApiResponse[ResponseControllerLeave]
	PilotJoin(req *RequestPilotJoin) *ApiResponse[ResponsePilotJoin]
	PilotLeave(req *RequestPilotLeave) *ApiResponse[ResponsePilotLeave]
	EditActivity(req *RequestEditActivity) *ApiResponse[ResponseEditActivity]
	EditPilotStatus(req *RequestEditPilotStatus) *ApiResponse[ResponseEditPilotStatus]
	EditActivityStatus(req *RequestEditActivityStatus) *ApiResponse[ResponseEditActivityStatus]
}

type RequestGetActivities struct {
	Time string `query:"time"`
}

type ResponseGetActivities struct {
	Items []*operation.Activity `json:"items"`
}

type RequestGetActivitiesPage struct {
	JwtHeader
	Page     int `query:"page_number"`
	PageSize int `query:"page_size"`
}

type ResponseGetActivitiesPage struct {
	Items    []*operation.Activity `json:"items"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

type RequestActivityInfo struct {
	ActivityId uint `param:"id"`
}

type ResponseActivityInfo operation.Activity

type RequestAddActivity struct {
	JwtHeader
	Cid int
	*operation.Activity
}

type ResponseAddActivity struct {
	*operation.Activity
}

type RequestDeleteActivity struct {
	JwtHeader
	ActivityId uint `param:"id"`
}

type ResponseDeleteActivity bool

type RequestControllerJoin struct {
	JwtHeader
	Rating     int
	ActivityId uint `param:"id"`
	FacilityId uint `param:"facility_id"`
}

type ResponseControllerJoin bool

type RequestControllerLeave struct {
	JwtHeader
	Cid        int
	ActivityId uint `param:"id"`
	FacilityId uint `param:"facility_id"`
}

type ResponseControllerLeave bool

type RequestPilotJoin struct {
	JwtHeader
	Cid          int
	ActivityId   uint   `param:"id"`
	Callsign     string `json:"callsign"`
	AircraftType string `json:"aircraft_type"`
}

type ResponsePilotJoin bool

type RequestPilotLeave struct {
	JwtHeader
	Cid        int
	ActivityId uint `param:"id"`
}

type ResponsePilotLeave bool

type RequestEditActivity struct {
	JwtHeader
	ActivityId       uint       `param:"id"`
	Title            *string    `json:"title"`
	ImageUrl         *string    `json:"image_url"`
	ActiveTime       *time.Time `json:"active_time"`
	DepartureAirport *string    `json:"departure_airport"`
	ArrivalAirport   *string    `json:"arrival_airport"`
	Route            *string    `json:"route"`
	Distance         *int       `json:"distance"`
	NOTAMS           *string    `json:"NOTAMS"`
}

type ResponseEditActivity bool

type RequestEditActivityStatus struct {
	JwtHeader
	ActivityId uint `param:"id"`
	Status     int  `json:"status"`
}

type ResponseEditActivityStatus bool

type RequestEditPilotStatus struct {
	JwtHeader
	ActivityId uint `param:"id"`
	Cid        int
	PilotId    uint `param:"pilot_id"`
	Status     int  `json:"status"`
}

type ResponseEditPilotStatus bool
