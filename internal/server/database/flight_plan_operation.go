package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/half-nothing/fsd-server/internal/utils"
	"gorm.io/gorm"
)

func GetFlightPlan(cid int) (*FlightPlan, error) {
	if config.Server.General.SimulatorServer {
		return nil, errors.New("simulator server not support flight plan store")
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	flightPlan := FlightPlan{}
	var err error
	err = database.WithContext(ctx).Where("cid=?", cid).First(&flightPlan).Error
	if err != nil {
		return nil, err
	}
	return &flightPlan, nil
}

func CreateFlightPlan(user *User, callsign string, flightPlanData []string) (*FlightPlan, error) {
	if len(flightPlanData) < 17 {
		return nil, fmt.Errorf("flight plan data is too short")
	}
	flightPlan := FlightPlan{
		Cid:      user.Cid,
		Callsign: callsign,
		Locked:   false,
		FromWeb:  false,
		Version:  0,
	}
	flightPlan.updateFlightPlanData(flightPlanData)
	// 模拟机服务器就不用写数据库了, 直接返回
	if config.Server.General.SimulatorServer {
		return &flightPlan, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	if err := database.WithContext(ctx).Save(&flightPlan).Error; err != nil {
		return nil, err
	}
	return &flightPlan, nil
}

func (flightPlan *FlightPlan) updateFlightPlanData(flightPlanData []string) {
	flightPlan.FlightType = flightPlanData[2]
	flightPlan.AircraftType = flightPlanData[3]
	flightPlan.Tas = utils.StrToInt(flightPlanData[4], 100)
	flightPlan.DepartureAirport = flightPlanData[5]
	flightPlan.DepartureTime = utils.StrToInt(flightPlanData[6], 0)
	flightPlan.AtcDepartureTime = utils.StrToInt(flightPlanData[7], 0)
	flightPlan.CruiseAltitude = flightPlanData[8]
	flightPlan.ArrivalAirport = flightPlanData[9]
	flightPlan.RouteTimeHour = flightPlanData[10]
	flightPlan.RouteTimeMinute = flightPlanData[11]
	flightPlan.FuelTimeHour = flightPlanData[12]
	flightPlan.FuelTimeMinute = flightPlanData[13]
	flightPlan.AlternateAirport = flightPlanData[14]
	flightPlan.Remarks = flightPlanData[15]
	flightPlan.Route = flightPlanData[16]
}

func (flightPlan *FlightPlan) UpdateFlightPlan(flightPlanData []string, atcEdit bool) error {
	if len(flightPlanData) < 17 {
		return fmt.Errorf("flight plan data is too short")
	}
	// 模拟机服务器只用更新内存中数据就行
	flightPlan.updateFlightPlanData(flightPlanData)
	if config.Server.General.SimulatorServer {
		return nil
	}
	return database.Transaction(func(tx *gorm.DB) error {
		// 在事务中重新加载最新数据
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		var current FlightPlan
		if err := tx.WithContext(ctx).Where("id = ?", flightPlan.ID).First(&current).Error; err != nil {
			return err
		}

		// 检查锁定状态（使用最新数据）
		if !atcEdit && current.Locked {
			return fmt.Errorf("flight plan locked")
		}

		// 准备更新
		updates := map[string]interface{}{
			"flight_type":        flightPlan.FlightType,
			"aircraft_type":      flightPlan.AircraftType,
			"tas":                flightPlan.Tas,
			"departure_airport":  flightPlan.DepartureAirport,
			"departure_time":     flightPlan.DepartureTime,
			"atc_departure_time": flightPlan.AtcDepartureTime,
			"cruise_altitude":    flightPlan.CruiseAltitude,
			"arrival_airport":    flightPlan.ArrivalAirport,
			"route_time_hour":    flightPlan.RouteTimeHour,
			"route_time_minute":  flightPlan.RouteTimeMinute,
			"fuel_time_hour":     flightPlan.FuelTimeHour,
			"fuel_time_minute":   flightPlan.FuelTimeMinute,
			"alternate_airport":  flightPlan.AlternateAirport,
			"remarks":            flightPlan.Remarks,
			"route":              flightPlan.Route,
			"version":            gorm.Expr("version + 1"),
		}

		result := tx.Model(flightPlan).
			Where("id = ? AND version = ?", flightPlan.ID, flightPlan.Version).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("flight plan modified by another request")
		}

		flightPlan.Version++
		return nil
	})
}

// UpdateCruiseAltitude
// 虽然理论上这个函数只能由ATC调用
// 但鬼知道会不会有用户闲的蛋疼手动给服务器发消息
// 所以还是多验证一点吧）
func (flightPlan *FlightPlan) UpdateCruiseAltitude(cruiseAltitude string, atcEdit bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	db := database.WithContext(ctx).Model(flightPlan)

	if !atcEdit {
		db = db.Where("locked = ?", false)
	}

	result := db.Where("id = ? AND version = ?", flightPlan.ID, flightPlan.Version).
		Updates(map[string]interface{}{
			"cruise_altitude": cruiseAltitude,
			"version":         gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("flight plan locked or modified by another request")
	}

	flightPlan.Version++
	return nil
}

func (flightPlan *FlightPlan) ToString(receiver string) string {
	return fmt.Sprintf("$FP%s:%s:%s:%s:%d:%s:%d:%d:%s:%s:%s:%s:%s:%s:%s:%s:%s\r\n",
		flightPlan.Callsign, receiver, flightPlan.FlightType, flightPlan.AircraftType, flightPlan.Tas,
		flightPlan.DepartureAirport, flightPlan.DepartureTime, flightPlan.AtcDepartureTime, flightPlan.CruiseAltitude,
		flightPlan.ArrivalAirport, flightPlan.RouteTimeHour, flightPlan.RouteTimeMinute, flightPlan.FuelTimeHour,
		flightPlan.FuelTimeMinute, flightPlan.AlternateAirport, flightPlan.Remarks, flightPlan.Route)
}
