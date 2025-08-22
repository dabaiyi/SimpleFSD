// Package service
package service

import (
	"errors"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/interfaces/fsd"
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
	. "github.com/half-nothing/fsd-server/internal/interfaces/service"
	"time"
)

type ActivityService struct {
	config            *c.HttpServerConfig
	userOperation     operation.UserOperationInterface
	activityOperation operation.ActivityOperationInterface
}

func NewActivityService(
	config *c.HttpServerConfig,
	userOperation operation.UserOperationInterface,
	activityOperation operation.ActivityOperationInterface,
) *ActivityService {
	return &ActivityService{
		config:            config,
		userOperation:     userOperation,
		activityOperation: activityOperation,
	}
}

var SuccessGetActivities = ApiStatus{StatusName: "GET_ACTIVITIES", Description: "成功获取活动", HttpCode: Ok}

func (activityService *ActivityService) GetActivities(req *RequestGetActivities) *ApiResponse[ResponseGetActivities] {
	firstDay, _ := time.Parse("2006-01", req.Time)
	nextMonth := firstDay.AddDate(0, 1, 0)
	lastDay := nextMonth.Add(-time.Second)
	activities, err := activityService.activityOperation.GetActivities(firstDay, lastDay)
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
	activity, res := CallDBFuncAndCheckError[operation.Activity, ResponseActivityInfo](func() (*operation.Activity, error) {
		return activityService.activityOperation.GetActivityById(req.ActivityId)
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
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityPublish) {
		return NewApiResponse[ResponseAddActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	req.Activity.Publisher = req.Cid
	err := activityService.activityOperation.SaveActivity(req.Activity)
	if err != nil {
		c.ErrorF("Error adding activity: %v", err)
		return NewApiResponse[ResponseAddActivity](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse[ResponseAddActivity](&SuccessAddActivity, Unsatisfied, &ResponseAddActivity{Activity: req.Activity})
}

var (
	SuccessDeleteActivity = ApiStatus{StatusName: "DELETE_ACTIVITY", Description: "成功删除活动", HttpCode: Ok}
)

func (activityService *ActivityService) DeleteActivity(req *RequestDeleteActivity) *ApiResponse[ResponseDeleteActivity] {
	if req.Permission <= 0 {
		return NewApiResponse[ResponseDeleteActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityDelete) {
		return NewApiResponse[ResponseDeleteActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	activity, err := activityService.activityOperation.GetActivityById(req.ActivityId)
	if err != nil {
		if errors.Is(err, operation.ErrActivityNotFound) {
			return NewApiResponse[ResponseDeleteActivity](&ErrActivityNotFound, Unsatisfied, nil)
		}
		return NewApiResponse[ResponseDeleteActivity](&ErrDatabaseFail, Unsatisfied, nil)
	}
	if err := activityService.activityOperation.DeleteActivity(activity); err != nil {
		return NewApiResponse[ResponseDeleteActivity](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseDeleteActivity(true)
	return NewApiResponse(&SuccessDeleteActivity, Unsatisfied, &data)
}

var (
	ErrRatingTooLow          = ApiStatus{StatusName: "RATING_TOO_LOW", Description: "管制权限不够", HttpCode: PermissionDenied}
	ErrFacilityAlreadyExist  = ApiStatus{StatusName: "FACILITY_ALREADY_EXIST", Description: "你不能同时报名两个以上的席位", HttpCode: Conflict}
	ErrFacilityAlreadySigned = ApiStatus{StatusName: "FACILITY_ALREADY_SIGNED", Description: "已有其他管制员报名", HttpCode: Conflict}
	SuccessSignFacility      = ApiStatus{StatusName: "SIGNED_FACILITY", Description: "报名成功", HttpCode: Ok}
)

func (activityService *ActivityService) ControllerJoin(req *RequestControllerJoin) *ApiResponse[ResponseControllerJoin] {
	if req.Rating <= fsd.Observer.Index() {
		return NewApiResponse[ResponseControllerJoin](&ErrRatingTooLow, Unsatisfied, nil)
	}
	user, res := CallDBFuncAndCheckError[operation.User, ResponseControllerJoin](func() (*operation.User, error) {
		return activityService.userOperation.GetUserByUid(req.Uid)
	})
	if res != nil {
		return res
	}
	facility, res := CallDBFuncAndCheckError[operation.ActivityFacility, ResponseControllerJoin](func() (*operation.ActivityFacility, error) {
		return activityService.activityOperation.GetFacilityById(req.FacilityId)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.SignFacilityController(facility, user)
	if err != nil {
		if errors.Is(err, operation.ErrRatingNotAllowed) {
			return NewApiResponse[ResponseControllerJoin](&ErrRatingTooLow, Unsatisfied, nil)
		}
		if errors.Is(err, operation.ErrFacilityAlreadyExists) {
			return NewApiResponse[ResponseControllerJoin](&ErrFacilityAlreadyExist, Unsatisfied, nil)
		}
		if errors.Is(err, operation.ErrFacilitySigned) {
			return NewApiResponse[ResponseControllerJoin](&ErrFacilityAlreadySigned, Unsatisfied, nil)
		}
		return NewApiResponse[ResponseControllerJoin](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseControllerJoin(true)
	return NewApiResponse(&SuccessSignFacility, Unsatisfied, &data)
}

var (
	ErrFacilityUnSigned   = ApiStatus{StatusName: "FACILITY_UNSIGNED", Description: "该席位尚未有人报名", HttpCode: Conflict}
	SuccessUnsignFacility = ApiStatus{StatusName: "UNSIGNED_FACILITY", Description: "成功取消报名", HttpCode: Ok}
)

func (activityService *ActivityService) ControllerLeave(req *RequestControllerLeave) *ApiResponse[ResponseControllerLeave] {
	facility, res := CallDBFuncAndCheckError[operation.ActivityFacility, ResponseControllerLeave](func() (*operation.ActivityFacility, error) {
		return activityService.activityOperation.GetFacilityById(req.FacilityId)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.UnsignFacilityController(facility, req.Cid)
	if err != nil {
		if errors.Is(err, operation.ErrFacilityNotSigned) {
			return NewApiResponse[ResponseControllerLeave](&ErrFacilityUnSigned, Unsatisfied, nil)
		}
		return NewApiResponse[ResponseControllerLeave](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseControllerLeave(true)
	return NewApiResponse(&SuccessUnsignFacility, Unsatisfied, &data)
}

var (
	ErrAlreadySigned      = ApiStatus{StatusName: "ALREADY_SIGNED", Description: "你已经报名该活动了", HttpCode: Conflict}
	ErrCallsignUsed       = ApiStatus{StatusName: "CALLSIGN_USED", Description: "呼号已被占用", HttpCode: Conflict}
	SuccessSignedActivity = ApiStatus{StatusName: "SIGNED_ACTIVITY", Description: "报名成功", HttpCode: Ok}
)

func (activityService *ActivityService) PilotJoin(req *RequestPilotJoin) *ApiResponse[ResponsePilotJoin] {
	activity, res := CallDBFuncAndCheckError[operation.Activity, ResponsePilotJoin](func() (*operation.Activity, error) {
		return activityService.activityOperation.GetOnlyActivityById(req.ActivityId)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.SignActivityPilot(activity, req.Cid, req.Callsign, req.AircraftType)
	if err != nil {
		if errors.Is(err, operation.ErrActivityAlreadySigned) {
			return NewApiResponse[ResponsePilotJoin](&ErrAlreadySigned, Unsatisfied, nil)
		}
		if errors.Is(err, operation.ErrCallsignAlreadyUsed) {
			return NewApiResponse[ResponsePilotJoin](&ErrCallsignUsed, Unsatisfied, nil)
		}
		return NewApiResponse[ResponsePilotJoin](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponsePilotJoin(true)
	return NewApiResponse(&SuccessSignedActivity, Unsatisfied, &data)
}

var (
	ErrNoSigned             = ApiStatus{StatusName: "NO_SIGNED", Description: "你还没有报名该活动", HttpCode: Conflict}
	SuccessUnsignedActivity = ApiStatus{StatusName: "UNSIGNED_ACTIVITY", Description: "取消报名成功", HttpCode: Ok}
)

func (activityService *ActivityService) PilotLeave(req *RequestPilotLeave) *ApiResponse[ResponsePilotLeave] {
	activity, res := CallDBFuncAndCheckError[operation.Activity, ResponsePilotLeave](func() (*operation.Activity, error) {
		return activityService.activityOperation.GetOnlyActivityById(req.ActivityId)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.UnsignActivityPilot(activity, req.Cid)
	if err != nil {
		if errors.Is(err, operation.ErrActivityUnsigned) {
			return NewApiResponse[ResponsePilotLeave](&ErrNoSigned, Unsatisfied, nil)
		}
		return NewApiResponse[ResponsePilotLeave](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponsePilotLeave(true)
	return NewApiResponse(&SuccessUnsignedActivity, Unsatisfied, &data)
}

var (
	SuccessEditActivity = ApiStatus{StatusName: "EDIT_ACTIVITY", Description: "修改活动成功", HttpCode: Ok}
)

func (activityService *ActivityService) EditActivity(req *RequestEditActivity) *ApiResponse[ResponseEditActivity] {
	if req.Title == nil && req.ImageUrl == nil && req.ActiveTime == nil && req.DepartureAirport == nil &&
		req.ArrivalAirport == nil && req.Route == nil && req.Distance == nil && req.NOTAMS == nil {
		return NewApiResponse[ResponseEditActivity](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseEditActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityEditContent) {
		return NewApiResponse[ResponseEditActivity](&ErrNoPermission, Unsatisfied, nil)
	}
	activity, res := CallDBFuncAndCheckError[operation.Activity, ResponseEditActivity](func() (*operation.Activity, error) {
		return activityService.activityOperation.GetOnlyActivityById(req.ActivityId)
	})
	if res != nil {
		return res
	}
	updateInfo := map[string]interface{}{}
	if req.Title != nil {
		updateInfo["title"] = req.Title
	}
	if req.ImageUrl != nil {
		updateInfo["image_url"] = req.ImageUrl
	}
	if req.ActiveTime != nil {
		updateInfo["active_time"] = req.ActiveTime
	}
	if req.DepartureAirport != nil {
		updateInfo["departure_airport"] = req.DepartureAirport
	}
	if req.ArrivalAirport != nil {
		updateInfo["arrival_airport"] = req.ArrivalAirport
	}
	if req.Route != nil {
		updateInfo["route"] = req.Route
	}
	if req.Distance != nil {
		updateInfo["distance"] = req.Distance
	}
	if req.NOTAMS != nil {
		updateInfo["notams"] = req.NOTAMS
	}
	err := activityService.activityOperation.UpdateActivityInfo(activity, updateInfo)
	if err != nil {
		return NewApiResponse[ResponseEditActivity](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseEditActivity(true)
	return NewApiResponse(&SuccessEditActivity, Unsatisfied, &data)
}

var SuccessEditActivityStatus = ApiStatus{StatusName: "EDIT_ACTIVITY_STATUS", Description: "成功修改活动状态", HttpCode: Ok}

func (activityService *ActivityService) EditActivityStatus(req *RequestEditActivityStatus) *ApiResponse[ResponseEditActivityStatus] {
	if req.Status < int(operation.Open) || req.Status > int(operation.Closed) {
		return NewApiResponse[ResponseEditActivityStatus](&ErrIllegalParam, Unsatisfied, nil)
	}

	status := operation.ActivityStatus(req.Status)

	if req.Permission <= 0 {
		return NewApiResponse[ResponseEditActivityStatus](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityEditState) {
		return NewApiResponse[ResponseEditActivityStatus](&ErrNoPermission, Unsatisfied, nil)
	}
	activity, res := CallDBFuncAndCheckError[operation.Activity, ResponseEditActivityStatus](func() (*operation.Activity, error) {
		return activityService.activityOperation.GetOnlyActivityById(req.ActivityId)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.SetActivityStatus(activity, status)
	if err != nil {
		return NewApiResponse[ResponseEditActivityStatus](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseEditActivityStatus(true)
	return NewApiResponse(&SuccessEditActivityStatus, Unsatisfied, &data)
}

var SuccessEditPilotsStatus = ApiStatus{StatusName: "EDIT_PILOTS_STATUS", Description: "成功修改活动机组状态", HttpCode: Ok}

func (activityService *ActivityService) EditPilotStatus(req *RequestEditPilotStatus) *ApiResponse[ResponseEditPilotStatus] {
	if req.Status < int(operation.Signed) || req.Status > int(operation.Landing) {
		return NewApiResponse[ResponseEditPilotStatus](&ErrIllegalParam, Unsatisfied, nil)
	}

	status := operation.ActivityPilotStatus(req.Status)

	if req.Permission <= 0 {
		return NewApiResponse[ResponseEditPilotStatus](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityEditState) {
		return NewApiResponse[ResponseEditPilotStatus](&ErrNoPermission, Unsatisfied, nil)
	}
	pilot, res := CallDBFuncAndCheckError[operation.ActivityPilot, ResponseEditPilotStatus](func() (*operation.ActivityPilot, error) {
		return activityService.activityOperation.GetActivityPilotById(req.ActivityId, req.Cid)
	})
	if res != nil {
		return res
	}
	err := activityService.activityOperation.SetActivityPilotStatus(pilot, status)
	if err != nil {
		return NewApiResponse[ResponseEditPilotStatus](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseEditPilotStatus(true)
	return NewApiResponse(&SuccessEditPilotsStatus, Unsatisfied, &data)
}
