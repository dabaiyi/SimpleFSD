// Package service
package service

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	. "github.com/half-nothing/fsd-server/internal/server/defination"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"time"
)

type ActivityService struct {
	config *c.HttpServerConfig
}

func NewActivityService(config *c.HttpServerConfig) *ActivityService {
	return &ActivityService{config}
}

var SuccessGetActivities = ApiStatus{StatusName: "GET_ACTIVITIES", Description: "成功获取活动", HttpCode: Ok}

func (activityService *ActivityService) GetActivities(req *RequestGetActivities) *ApiResponse[ResponseGetActivities] {
	firstDay, _ := time.Parse("2006-01", req.Time)
	nextMonth := firstDay.AddDate(0, 1, 0)
	lastDay := nextMonth.Add(-time.Second)
	activities, err := database.GetActivities(firstDay, lastDay)
	if err != nil {
		return NewApiResponse[ResponseGetActivities](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse[ResponseGetActivities](&SuccessGetActivities, Unsatisfied, &ResponseGetActivities{Items: activities})
}

var SuccessGetActivityInfo = ApiStatus{StatusName: "GET_ACTIVITY_INFO", Description: "成功获取活动信息", HttpCode: Ok}

func (activityService *ActivityService) GetActivityInfo(req *RequestActivityInfo) *ApiResponse[ResponseActivityInfo] {
	if req.ActivityId <= 0 {
		return NewApiResponse[ResponseActivityInfo](&ErrIllegalParam, Unsatisfied, nil)
	}
	activity, res := CallDBFuncAndCheckError[database.Activity, ResponseActivityInfo](func() (*database.Activity, error) {
		return database.GetActivityById(req.ActivityId)
	})
	if res != nil {
		return res
	}
	return NewApiResponse(&SuccessGetActivityInfo, Unsatisfied, (*ResponseActivityInfo)(activity))
}

var SuccessAddActivity = ApiStatus{StatusName: "ADD_ACTIVITY", Description: "成功添加活动", HttpCode: Ok}

func (activityService *ActivityService) AddActivity(req *RequestAddActivity) *ApiResponse[ResponseAddActivity] {
	if req.Permission <= 0 {
		return NewApiResponse[ResponseAddActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := Permission(req.Permission)
	if !permission.HasPermission(ActivityPublish) {
		return NewApiResponse[ResponseAddActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	req.Activity.Publisher = req.Cid
	err := req.Activity.Save()
	if err != nil {
		c.ErrorF("Error adding activity: %v", err)
		return NewApiResponse[ResponseAddActivity](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse[ResponseAddActivity](&SuccessAddActivity, Unsatisfied, &ResponseAddActivity{Activity: req.Activity})
}

func (activityService *ActivityService) DeleteActivity(req *RequestDeleteActivity) *ApiResponse[ResponseDeleteActivity] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) EditActivity(req *RequestEditActivity) *ApiResponse[ResponseEditActivity] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) ControllerJoin(req *RequestControllerJoin) *ApiResponse[ResponseControllerJoin] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) ControllerLeave(req *RequestControllerLeave) *ApiResponse[ResponseControllerLeave] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) PilotJoin(req *RequestPilotJoin) *ApiResponse[ResponsePilotJoin] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) PilotLeave(req *RequestPilotLeave) *ApiResponse[ResponsePilotLeave] {
	//TODO implement me
	panic("implement me")
}

func (activityService *ActivityService) EditPilotStatus(req *RequestEditPilotStatus) *ApiResponse[ResponseEditPilotStatus] {
	//TODO implement me
	panic("implement me")
}
