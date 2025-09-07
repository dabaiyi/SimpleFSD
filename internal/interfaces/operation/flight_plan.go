// Package operation
package operation

import (
	"errors"
)

var (
	ErrFlightPlanNotFound     = errors.New("flight plan not found")
	ErrSimulatorServer        = errors.New("simulator fsd_server not support flight plan store")
	ErrFlightPlanDataTooShort = errors.New("flight plan data is too short")
	ErrFlightPlanExists       = errors.New("flight plan already exists")
	ErrFlightPlanLocked       = errors.New("flight plan locked")
)

// FlightPlanOperationInterface 飞行计划操作接口定义
type FlightPlanOperationInterface interface {
	// GetFlightPlanByCid 通过用户cid获取飞行计划, 当err为nil时返回值flightPlan有效
	GetFlightPlanByCid(cid int) (flightPlan *FlightPlan, err error)
	// UpsertFlightPlan 创建或更新飞行计划, 当err为nil时返回值flightPlan有效
	UpsertFlightPlan(user *User, callsign string, flightPlanData []string) (flightPlan *FlightPlan, err error)
	// UpdateFlightPlanData 更新飞行计划(不提交数据库)
	UpdateFlightPlanData(flightPlan *FlightPlan, flightPlanData []string)
	// UpdateFlightPlan 更新飞行计划(提交数据库), 当err为nil时更新成功
	UpdateFlightPlan(flightPlan *FlightPlan, flightPlanData []string, atcEdit bool) (err error)
	// UpdateCruiseAltitude 更新巡航高度, 当err为nil时更新成功
	UpdateCruiseAltitude(flightPlan *FlightPlan, cruiseAltitude string) (err error)
	// ToString 将飞行计划转换为ES和Swift可识别的形式
	ToString(flightPlan *FlightPlan, receiver string) (str string)
}
