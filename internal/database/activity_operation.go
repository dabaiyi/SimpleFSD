package database

import (
	"context"
	"fmt"
	. "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"gorm.io/gorm"
	"time"
)

type ActivityOperation struct {
	db *gorm.DB
}

func NewActivityOperation(db *gorm.DB) *ActivityOperation {
	return &ActivityOperation{db: db}
}

func (activityOperation *ActivityOperation) NewActivity(user *User, title string, imageUrl string, activeTime time.Time, dep string, arr string, route string, distance int, notams string) (activity *Activity) {
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

func (activityOperation *ActivityOperation) NewActivityFacility(activity *Activity, rating int, callsign string, frequency float64) (activityFacility *ActivityFacility) {
	return &ActivityFacility{
		ActivityId: activity.ID,
		MinRating:  rating,
		Callsign:   callsign,
		Frequency:  fmt.Sprintf("%.3f", frequency),
	}
}

func (activityOperation *ActivityOperation) NewActivityAtc(activity *Activity, facility *ActivityFacility) (activityAtc *ActivityATC) {
	return &ActivityATC{
		ActivityId: activity.ID,
		FacilityId: facility.ID,
		Cid:        0,
	}
}

func (activityOperation *ActivityOperation) NewActivityPilot(activity *Activity, user *User, callsign string, aircraftType string) (activityPilot *ActivityPilot) {
	return &ActivityPilot{
		ActivityId:   activity.ID,
		Cid:          user.Cid,
		Callsign:     callsign,
		AircraftType: aircraftType,
		Status:       int(Signed),
	}
}

func (activityOperation *ActivityOperation) GetActivities(startDay, endDay time.Time) (activities []*Activity, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err = activityOperation.db.WithContext(ctx).Where("active_time between ? and ?", startDay, endDay).Find(&activities).Error
	return
}

func (activityOperation *ActivityOperation) GetActivityById(id uint) (activity *Activity, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err = activityOperation.db.WithContext(ctx).
		Preload("Facilities").
		Preload("Pilots").
		Preload("Controllers").
		Where("id = ?", id).
		First(&activity).
		Error
	return
}

func (activityOperation *ActivityOperation) SaveActivity(activity *Activity) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	return activityOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(activity).Error
	})
}

func (activityOperation *ActivityOperation) DeleteActivity(activity *Activity) (err error) {
	return activityOperation.db.Transaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", activity.ID).
			Delete(&ActivityPilot{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity pilots: %w", err)
		}

		if err := tx.WithContext(ctx).
			Where("activity_id = ?", activity.ID).
			Delete(&ActivityATC{}).Error; err != nil {
			return fmt.Errorf("fail to delete activity atcs: %w", err)
		}

		if err := tx.WithContext(ctx).Delete(activity).Error; err != nil {
			return fmt.Errorf("fail to delete activity: %w", err)
		}

		return nil
	})
}

func (activityOperation *ActivityOperation) GetActivityPilots(activity *Activity) (pilots []*ActivityPilot, err error) {
	if activity.Pilots != nil {
		return activity.Pilots, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err = activityOperation.db.WithContext(ctx).
		Where("activity_id = ?", activity.ID).
		Find(&pilots).Error
	return
}

func (activityOperation *ActivityOperation) GetActivityATCs(activity *Activity) (atcs []*ActivityATC, err error) {
	if activity.Controllers != nil {
		return activity.Controllers, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err = activityOperation.db.WithContext(ctx).
		Where("activity_id = ?", activity.ID).
		Find(&atcs).Error
	return
}

func (activityOperation *ActivityOperation) GetActivityFacilities(activity *Activity) (facilities []*ActivityFacility, err error) {
	if activity.Facilities != nil {
		return activity.Facilities, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err = activityOperation.db.WithContext(ctx).
		Where("activity_id = ?", activity.ID).
		Find(&facilities).Error
	return
}

func (activityOperation *ActivityOperation) SetActivityStatus(activity *Activity, status ActivityStatus) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	return activityOperation.db.WithContext(ctx).Model(activity).Update("status", int(status)).Error
}

func (activityOperation *ActivityOperation) SetActivityStatusOpen(activity *Activity) (err error) {
	return activityOperation.SetActivityStatus(activity, Open)
}

func (activityOperation *ActivityOperation) SetActivityStatusActive(activity *Activity) (err error) {
	return activityOperation.SetActivityStatus(activity, InActive)
}

func (activityOperation *ActivityOperation) SetActivityStatusClosed(activity *Activity) (err error) {
	return activityOperation.SetActivityStatus(activity, Closed)
}

func (activityOperation *ActivityOperation) SaveActivityPilot(activityPilot *ActivityPilot) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Save(activityPilot).Error
}

func (activityOperation *ActivityOperation) SetActivityPilotStatus(activityPilot *ActivityPilot, status ActivityPilotStatus) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Model(activityPilot).Update("status", int(status)).Error
}

func (activityOperation *ActivityOperation) SetActivityPilotStatusSigned(activityPilot *ActivityPilot) (err error) {
	return activityOperation.SetActivityPilotStatus(activityPilot, Signed)
}

func (activityOperation *ActivityOperation) SetActivityPilotStatusClearance(activityPilot *ActivityPilot) (err error) {
	return activityOperation.SetActivityPilotStatus(activityPilot, Clearance)
}

func (activityOperation *ActivityOperation) SetActivityPilotStatusTakeoff(activityPilot *ActivityPilot) (err error) {
	return activityOperation.SetActivityPilotStatus(activityPilot, Takeoff)
}

func (activityOperation *ActivityOperation) SetActivityPilotStatusLanding(activityPilot *ActivityPilot) (err error) {
	return activityOperation.SetActivityPilotStatus(activityPilot, Landing)
}

func (activityOperation *ActivityOperation) SaveActivityAtc(activityAtc *ActivityATC) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Save(activityAtc).Error
}

func (activityOperation *ActivityOperation) SetActivityFacility(activity *ActivityFacility, user *User) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Model(activity).Update("cid", user.Cid).Error
}
