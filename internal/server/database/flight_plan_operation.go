package database

import (
	"context"
	"fmt"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
)

func GetFlightPlan(cid int, callsign string) (*FlightPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	flightPlan := FlightPlan{}
	var err error
	if config.SimulatorServer {
		err = database.WithContext(ctx).Where("callsign=?", callsign).First(&flightPlan).Error
	} else {
		err = database.WithContext(ctx).Where("cid=?", cid).First(&flightPlan).Error
	}
	if err != nil {
		return nil, err
	}
	return &flightPlan, nil
}

func CreateFlightPlan(user *User, callsign string, flightPlanData []string) (*FlightPlan, error) {
	flightPlan := FlightPlan{
		Cid:              user.Cid,
		Callsign:         callsign,
		FlightType:       flightPlanData[2],
		AircraftType:     flightPlanData[3],
		Tas:              utils.StrToInt(flightPlanData[4], 100),
		DepartureAirport: flightPlanData[5],
		DepartureTime:    utils.StrToInt(flightPlanData[6], 0),
		AtcDepartureTime: utils.StrToInt(flightPlanData[7], 0),
		CruiseAltitude:   flightPlanData[8],
		ArrivalAirport:   flightPlanData[9],
		RouteTimeHour:    flightPlanData[10],
		RouteTimeMinute:  flightPlanData[11],
		FuelTimeHour:     flightPlanData[12],
		FuelTimeMinute:   flightPlanData[13],
		AlternateAirport: flightPlanData[14],
		Remarks:          flightPlanData[15],
		Route:            flightPlanData[16],
		Locked:           false,
		FromWeb:          false,
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	if err := database.WithContext(ctx).Save(&flightPlan).Error; err != nil {
		return nil, err
	}
	return &flightPlan, nil
}

func (flightPlan *FlightPlan) Lock() bool {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(flightPlan).Update("locked", true).Error
	return err == nil
}

func (flightPlan *FlightPlan) Unlock() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(flightPlan).Update("locked", false).Error
	return err
}

func (flightPlan *FlightPlan) UpdateFlightPlan(flightPlanData []string, atcEdit bool) error {
	if !atcEdit && flightPlan.Locked {
		return nil
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Save(&flightPlan).Error
	return err
}

func (flightPlan *FlightPlan) UpdateCruiseAltitude(cruiseAltitude string, atcEdit bool) error {
	if !atcEdit && flightPlan.Locked {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Model(flightPlan).Update("cruise_altitude", cruiseAltitude).Error
	return err
}

func (flightPlan *FlightPlan) ToString(receiver string) string {
	return fmt.Sprintf("$FP%s:%s:%s:%s:%d:%s:%d:%d:%s:%s:%s:%s:%s:%s:%s:%s:%s\r\n",
		flightPlan.Callsign, receiver, flightPlan.FlightType, flightPlan.AircraftType, flightPlan.Tas,
		flightPlan.DepartureAirport, flightPlan.DepartureTime, flightPlan.AtcDepartureTime, flightPlan.CruiseAltitude,
		flightPlan.ArrivalAirport, flightPlan.RouteTimeHour, flightPlan.RouteTimeMinute, flightPlan.FuelTimeHour,
		flightPlan.FuelTimeMinute, flightPlan.AlternateAirport, flightPlan.Remarks, flightPlan.Route)
}
