package database

import (
	"context"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
)

func GetFlightPlan(callsign string) (*FlightPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	flightPlan := FlightPlan{}
	if err := database.WithContext(ctx).Where("callsign = ?", callsign).First(&flightPlan).Error; err != nil {
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

func (flightPlan *FlightPlan) Unlock() bool {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(flightPlan).Update("locked", false).Error
	return err == nil
}

func (flightPlan *FlightPlan) UpdateFlightPlan(flightPlanData []string) bool {
	if flightPlan.Locked {
		return false
	}
	flightPlan.FlightType = flightPlanData[2]
	flightPlan.AircraftType = flightPlanData[3]
	flightPlan.Tas = utils.StrToInt(flightPlanData[4], 100)
	flightPlan.DepartureAirport = flightPlanData[5]
	flightPlan.DepartureTime = utils.StrToInt(flightPlanData[6], 0)
	flightPlan.AtcDepartureTime = utils.StrToInt(flightPlanData[7], 0)
	flightPlan.CruiseAltitude = flightPlanData[8]
	flightPlan.ArrivalAirport = flightPlanData[9]
	flightPlan.RouteTime = utils.StrToInt(flightPlanData[10]+flightPlanData[11], 0)
	flightPlan.FuelTime = utils.StrToInt(flightPlanData[12]+flightPlanData[13], 0)
	flightPlan.AlternateAirport = flightPlanData[14]
	flightPlan.Remarks = flightPlanData[15]
	flightPlan.Route = flightPlanData[16]
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Save(&flightPlan).Error
	return err == nil
}
