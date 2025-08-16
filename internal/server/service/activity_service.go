// Package service
package service

import . "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"

type ActivityService struct {
}

func (activityService ActivityService) GetActivities(req *RequestGetActivities) *ApiResponse[ResponseGetActivities] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) GetActivityInfo(req *RequestActivityInfo) *ApiResponse[ResponseActivityInfo] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) AddActivity(req *RequestAddActivity) *ApiResponse[ResponseAddActivity] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) DeleteActivity(req *RequestDeleteActivity) *ApiResponse[ResponseDeleteActivity] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) EditActivity(req *RequestEditActivity) *ApiResponse[ResponseEditActivity] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) ControllerJoin(req *RequestControllerJoin) *ApiResponse[ResponseControllerJoin] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) ControllerLeave(req *RequestControllerLeave) *ApiResponse[ResponseControllerLeave] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) PilotJoin(req *RequestPilotJoin) *ApiResponse[ResponsePilotJoin] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) PilotLeave(req *RequestPilotLeave) *ApiResponse[ResponsePilotLeave] {
	//TODO implement me
	panic("implement me")
}

func (activityService ActivityService) EditPilotStatus(req *RequestEditPilotStatus) *ApiResponse[ResponseEditPilotStatus] {
	//TODO implement me
	panic("implement me")
}
