// Package database
package database

import (
	"context"
	"fmt"
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

func (user *User) NewActivity(title string, imageUrl string, activeTime time.Time, dep string,
	arr string, route string, distance int, notams string) *Activity {
	return &Activity{
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

func (ac *Activity) NewActivityAtc(rating int, callsign string, frequency float64) *ActivityATC {
	return &ActivityATC{
		ActivityId: ac.ID,
		Cid:        0,
		MinRating:  rating,
		Callsign:   callsign,
		Frequency:  fmt.Sprintf("%.3f", frequency),
	}
}

func (ac *Activity) NewActivityPilot(user *User, callsign string, aircraftType string) *ActivityPilot {
	return &ActivityPilot{
		ActivityId:   ac.ID,
		Cid:          user.Cid,
		Callsign:     callsign,
		AircraftType: aircraftType,
		Status:       int(Signed),
	}
}

func (ac *Activity) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(ac).Error
	return err
}

func (ac *Activity) Delete() error {
	return database.Transaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", ac.ID).
			Delete(&ActivityPilot{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity pilots: %w", err)
		}

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", ac.ID).
			Delete(&ActivityATC{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity atcs: %w", err)
		}

		if err := tx.WithContext(ctx).Delete(ac).Error; err != nil {
			return fmt.Errorf("fail to delete activity: %w", err)
		}

		return nil
	})
}

func (ac *Activity) GetPilots() ([]ActivityPilot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var pilots []ActivityPilot
	err := database.WithContext(ctx).
		Where("activity_id = ?", ac.ID).
		Find(&pilots).Error
	return pilots, err
}

func (ac *Activity) GetATCs() ([]ActivityATC, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var atcs []ActivityATC
	err := database.WithContext(ctx).
		Where("activity_id = ?", ac.ID).
		Find(&atcs).Error
	return atcs, err
}

func (ac *Activity) setStatus(status ActivityStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Model(ac).Update("status", int(status)).Error
	return err
}

func (ac *Activity) SetOpen() error {
	return ac.setStatus(Open)
}

func (ac *Activity) SetInActive() error {
	return ac.setStatus(InActive)
}

func (ac *Activity) SetClosed() error {
	return ac.setStatus(Closed)
}

func (acp *ActivityPilot) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(acp).Error
	return err
}

func (acp *ActivityPilot) setStatus(status ActivityPilotStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Model(acp).Update("status", int(status)).Error
	return err
}

func (acp *ActivityPilot) SetSigned() error {
	return acp.setStatus(Signed)
}

func (acp *ActivityPilot) SetClearance() error {
	return acp.setStatus(Clearance)
}

func (acp *ActivityPilot) SetTakeoff() error {
	return acp.setStatus(Takeoff)
}

func (acp *ActivityPilot) SetLanding() error {
	return acp.setStatus(Landing)
}

func (aca *ActivityATC) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(aca).Error
	return err
}

func (aca *ActivityATC) Cancel() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(aca).Update("cid", 0).Error
	return err
}

func (aca *ActivityATC) SetAtc(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	err := database.WithContext(ctx).Model(aca).Update("cid", user.Cid).Error
	return err
}
