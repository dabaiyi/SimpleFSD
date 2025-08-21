// Package database
package database

import (
	"context"
	"fmt"
	database2 "github.com/half-nothing/fsd-server/internal/server/defination/database"
	"gorm.io/gorm"
	"time"
)

type ActivityStatus int

const (
	Open     ActivityStatus = iota // 报名中
	InActive                       // 活动中
	Closed                         // 已结束
)

type ActivityPilotStatus int

const (
	Signed    ActivityPilotStatus = iota // 已报名
	Clearance                            // 已放行
	Takeoff                              // 已起飞
	Landing                              // 已落地
)

func (user *database2.User) NewActivity(title string, imageUrl string, activeTime time.Time, dep string,
	arr string, route string, distance int, notams string) *database2.Activity {
	return &database2.Activity{
		Publisher:        user.Cid,
		Title:            title,
		ImageUrl:         imageUrl,
		ActiveTime:       activeTime,
		DepartureAirport: dep,
		ArrivalAirport:   arr,
		Route:            route,
		Distance:         distance,
		Status:           int(Open),
		NOTAMS:           notams,
	}
}

func (ac *database2.Activity) NewActivityFacility(rating int, callsign string, frequency float64) *database2.ActivityFacility {
	return &database2.ActivityFacility{
		ActivityId: ac.ID,
		MinRating:  rating,
		Callsign:   callsign,
		Frequency:  fmt.Sprintf("%.3f", frequency),
	}
}

func (ac *database2.Activity) NewActivityAtc(facility *database2.ActivityFacility) *database2.ActivityATC {
	return &database2.ActivityATC{
		ActivityId: ac.ID,
		FacilityId: facility.ID,
		Cid:        0,
	}
}

func (ac *database2.Activity) NewActivityPilot(user *database2.User, callsign string, aircraftType string) *database2.ActivityPilot {
	return &database2.ActivityPilot{
		ActivityId:   ac.ID,
		Cid:          user.Cid,
		Callsign:     callsign,
		AircraftType: aircraftType,
		Status:       int(Signed),
	}
}

func GetActivities(startDay, endDay time.Time) ([]*database2.Activity, error) {
	activities := make([]*database2.Activity, 0)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Where("active_time between ? and ?", startDay, endDay).Find(&activities).Error

	return activities, err
}

func GetActivityById(id uint) (*database2.Activity, error) {
	activity := &database2.Activity{}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).
		Preload("Facilities").
		Preload("Pilots").
		Preload("Controllers").
		Where("id = ?", id).
		First(&activity).
		Error

	return activity, err
}

func (ac *database2.Activity) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(ac).Error
	})
}

func (ac *database2.Activity) Delete() error {
	return database.Transaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", ac.ID).
			Delete(&database2.ActivityPilot{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity pilots: %w", err)
		}

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", ac.ID).
			Delete(&database2.ActivityATC{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity atcs: %w", err)
		}

		if err := tx.WithContext(ctx).Delete(ac).Error; err != nil {
			return fmt.Errorf("fail to delete activity: %w", err)
		}

		return nil
	})
}

func (ac *database2.Activity) GetPilots() ([]database2.ActivityPilot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var pilots []database2.ActivityPilot
	err := database.WithContext(ctx).
		Where("activity_id = ?", ac.ID).
		Find(&pilots).Error
	return pilots, err
}

func (ac *database2.Activity) GetATCs() ([]database2.ActivityATC, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var atcs []database2.ActivityATC
	err := database.WithContext(ctx).
		Where("activity_id = ?", ac.ID).
		Find(&atcs).Error
	return atcs, err
}

func (ac *database2.Activity) setStatus(status ActivityStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Model(ac).Update("status", int(status)).Error
	return err
}

func (ac *database2.Activity) SetOpen() error {
	return ac.setStatus(Open)
}

func (ac *database2.Activity) SetInActive() error {
	return ac.setStatus(InActive)
}

func (ac *database2.Activity) SetClosed() error {
	return ac.setStatus(Closed)
}

func (acp *database2.ActivityPilot) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(acp).Error
	return err
}

func (acp *database2.ActivityPilot) setStatus(status ActivityPilotStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Model(acp).Update("status", int(status)).Error
	return err
}

func (acp *database2.ActivityPilot) SetSigned() error {
	return acp.setStatus(Signed)
}

func (acp *database2.ActivityPilot) SetClearance() error {
	return acp.setStatus(Clearance)
}

func (acp *database2.ActivityPilot) SetTakeoff() error {
	return acp.setStatus(Takeoff)
}

func (acp *database2.ActivityPilot) SetLanding() error {
	return acp.setStatus(Landing)
}

func (aca *database2.ActivityATC) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(aca).Error
	return err
}

func (aca *database2.ActivityATC) Cancel() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(aca).Update("cid", 0).Error
	return err
}

func (aca *database2.ActivityATC) SetAtc(user *database2.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(aca).Update("cid", user.Cid).Error
	return err
}
