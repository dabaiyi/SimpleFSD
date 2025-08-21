package database

import (
	"context"
	"errors"
	"fmt"
	database2 "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"github.com/half-nothing/fsd-server/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrFlightPlanNotFound     = errors.New("flight plan not found")
	ErrSimulatorServer        = errors.New("simulator fsd_server not support flight plan store")
	ErrFlightPlanDataTooShort = fmt.Errorf("flight plan data is too short")
	ErrFlightPlanExists       = errors.New("flight plan already exists")
	ErrFlightPlanLocked       = errors.New("flight plan locked")
)

func GetFlightPlan(cid int) (*database2.FlightPlan, error) {
	if config.Server.General.SimulatorServer {
		return nil, ErrSimulatorServer
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	flightPlan := database2.FlightPlan{}
	var err error
	err = database.WithContext(ctx).Where("cid = ?", cid).First(&flightPlan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFlightPlanNotFound
	} else if err != nil {
		return nil, err
	}
	return &flightPlan, nil
}

func UpsertFlightPlan(user *database2.User, callsign string, flightPlanData []string) (*database2.FlightPlan, error) {
	if len(flightPlanData) < 17 {
		return nil, ErrFlightPlanDataTooShort
	}
	// 再次检查一遍防止重复创建
	flightPlan, err := GetFlightPlan(user.Cid)
	if err != nil {
		flightPlan = &database2.FlightPlan{
			Cid:      user.Cid,
			Callsign: callsign,
			Locked:   false,
			FromWeb:  false,
		}
	}
	flightPlan.updateFlightPlanData(flightPlanData)
	// 模拟机服务器就不用写数据库了, 直接返回
	if config.Server.General.SimulatorServer {
		return flightPlan, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err = database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(flightPlan).Error; errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrFlightPlanExists
		} else if err != nil {
			return err
		}
		return nil
	})
	return flightPlan, err
}

func (flightPlan *database2.FlightPlan) updateFlightPlanData(flightPlanData []string) {
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

func (flightPlan *database2.FlightPlan) UpdateFlightPlan(flightPlanData []string, atcEdit bool) error {
	if len(flightPlanData) < 17 {
		return ErrFlightPlanDataTooShort
	}
	// 模拟机服务器只用更新内存中数据就行
	flightPlan.updateFlightPlanData(flightPlanData)
	if config.Server.General.SimulatorServer {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Save(flightPlan).Error
	})
}

// UpdateCruiseAltitude
// 虽然理论上这个函数只能由ATC调用
// 但鬼知道会不会有用户闲的蛋疼手动给服务器发消息
// 所以还是多验证一点吧）
func (flightPlan *database2.FlightPlan) UpdateCruiseAltitude(cruiseAltitude string, atcEdit bool) error {
	flightPlan.CruiseAltitude = cruiseAltitude
	if config.Server.General.SimulatorServer {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	if !atcEdit && flightPlan.Locked {
		return ErrFlightPlanLocked
	}

	result := database.WithContext(ctx).Model(flightPlan).Where("locked = ?", false).Update("cruise_altitude", cruiseAltitude)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return ErrFlightPlanNotFound
	} else if result.RowsAffected == 0 {
		return ErrFlightPlanLocked
	} else if result.Error != nil {
		return result.Error
	}
	return nil
}

func (flightPlan *database2.FlightPlan) ToString(receiver string) string {
	return fmt.Sprintf("$FP%s:%s:%s:%s:%d:%s:%d:%d:%s:%s:%s:%s:%s:%s:%s:%s:%s\r\n",
		flightPlan.Callsign, receiver, flightPlan.FlightType, flightPlan.AircraftType, flightPlan.Tas,
		flightPlan.DepartureAirport, flightPlan.DepartureTime, flightPlan.AtcDepartureTime, flightPlan.CruiseAltitude,
		flightPlan.ArrivalAirport, flightPlan.RouteTimeHour, flightPlan.RouteTimeMinute, flightPlan.FuelTimeHour,
		flightPlan.FuelTimeMinute, flightPlan.AlternateAirport, flightPlan.Remarks, flightPlan.Route)
}
