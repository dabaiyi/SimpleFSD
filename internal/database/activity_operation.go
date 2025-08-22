package database

import (
	"context"
	"errors"
	"fmt"
	. "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ActivityOperation struct {
	db           *gorm.DB
	queryTimeout time.Duration
}

func NewActivityOperation(db *gorm.DB, queryTimeout time.Duration) *ActivityOperation {
	return &ActivityOperation{db: db, queryTimeout: queryTimeout}
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

func (activityOperation *ActivityOperation) NewActivityAtc(facility *ActivityFacility, user *User) (activityAtc *ActivityATC) {
	return &ActivityATC{
		ActivityId: facility.ActivityId,
		FacilityId: facility.ID,
		Cid:        user.Cid,
	}
}

func (activityOperation *ActivityOperation) NewActivityPilot(activity *Activity, cid int, callsign string, aircraftType string) (activityPilot *ActivityPilot) {
	return &ActivityPilot{
		ActivityId:   activity.ID,
		Cid:          cid,
		Callsign:     callsign,
		AircraftType: aircraftType,
		Status:       int(Signed),
	}
}

func (activityOperation *ActivityOperation) GetActivities(startDay, endDay time.Time) (activities []*Activity, err error) {
	activities = make([]*Activity, 0)
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	err = activityOperation.db.WithContext(ctx).Where("active_time between ? and ?", startDay, endDay).Find(&activities).Error
	return
}

func (activityOperation *ActivityOperation) GetActivityById(id uint) (activity *Activity, err error) {
	activity = &Activity{}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	err = activityOperation.db.WithContext(ctx).
		Preload("Facilities").
		Preload("Pilots").
		Preload("Controllers").
		Where("id = ?", id).
		First(activity).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrActivityNotFound
	}
	return
}

func (activityOperation *ActivityOperation) GetOnlyActivityById(id uint) (activity *Activity, err error) {
	activity = &Activity{}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	err = activityOperation.db.WithContext(ctx).Where("id = ?", id).First(activity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrActivityNotFound
	}
	return
}

func (activityOperation *ActivityOperation) SaveActivity(activity *Activity) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(activity).Error
	})
}

func (activityOperation *ActivityOperation) DeleteActivity(activity *Activity) (err error) {
	return activityOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).Transaction(func(tx *gorm.DB) error {
		ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
		defer cancel()
		// 软删除不用同步删除外键
		if err := tx.WithContext(ctx).Delete(activity).Error; err != nil {
			return fmt.Errorf("fail to delete activity: %w", err)
		}
		return nil
	})
}

func (activityOperation *ActivityOperation) SetActivityStatus(activity *Activity, status ActivityStatus) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Model(activity).Update("status", int(status)).Error
}

func (activityOperation *ActivityOperation) SetActivityPilotStatus(activityPilot *ActivityPilot, status ActivityPilotStatus) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Model(activityPilot).Update("status", int(status)).Error
}

func (activityOperation *ActivityOperation) GetFacilityById(facilityId uint) (facility *ActivityFacility, err error) {
	facility = &ActivityFacility{}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	err = activityOperation.db.WithContext(ctx).Preload("Controller").Where("id = ?", facilityId).First(facility).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrFacilityNotFound
	}
	return
}

func (activityOperation *ActivityOperation) GetActivityPilotById(activityId uint, cid int) (pilot *ActivityPilot, err error) {
	pilot = &ActivityPilot{}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	err = activityOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("activity_id = ? and cid = ?", activityId, cid).First(pilot).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrActivityUnsigned
		}
		return err
	})
	return
}

func (activityOperation *ActivityOperation) SignFacilityController(facility *ActivityFacility, user *User) (err error) {
	if user.Rating <= facility.MinRating {
		return ErrRatingNotAllowed
	}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		controller := &ActivityATC{}
		tx.Select("id").Where("activity_id = ? and cid = ?", facility.ActivityId, user.Cid).First(controller)
		if controller.ID != 0 {
			return ErrFacilityAlreadyExists
		}
		activityController := activityOperation.NewActivityAtc(facility, user)
		err := tx.Create(activityController).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrFacilitySigned
		}
		return err
	})
}

func (activityOperation *ActivityOperation) UnsignFacilityController(facility *ActivityFacility, cid int) (err error) {
	if facility.Controller == nil {
		return ErrFacilityNotSigned
	}
	if facility.Controller.Cid != cid {
		return ErrFacilityNotYourSign
	}
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		controller := &ActivityATC{}
		tx.Select("id").Where("activity_id = ? and facility_id = ? and cid = ?", facility.ActivityId, facility.ID, cid).First(controller)
		if controller.ID == 0 {
			return ErrFacilityNotSigned
		}
		return tx.Delete(controller).Error
	})
}

func (activityOperation *ActivityOperation) SignActivityPilot(activity *Activity, cid int, callsign string, aircraftType string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pilot := &ActivityPilot{}
		tx.Select("id", "cid", "callsign").Where("activity_id = ? and (cid = ? or callsign = ?)", activity.ID, cid, callsign).First(pilot)
		if pilot.ID != 0 {
			if pilot.Cid == cid {
				return ErrActivityAlreadySigned
			}
			return ErrCallsignAlreadyUsed
		}
		activityPilot := activityOperation.NewActivityPilot(activity, cid, callsign, aircraftType)
		err := tx.Create(activityPilot).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrActivityAlreadySigned
		}
		return err
	})
}

func (activityOperation *ActivityOperation) UnsignActivityPilot(activity *Activity, cid int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pilot := &ActivityPilot{}
		tx.Select("id").Where("activity_id = ? and cid = ?", activity.ID, cid).First(pilot)
		if pilot.ID == 0 {
			return ErrActivityUnsigned
		}
		return tx.Delete(pilot).Error
	})
}

func (activityOperation *ActivityOperation) UpdateActivityInfo(activity *Activity, updateInfo map[string]interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), activityOperation.queryTimeout)
	defer cancel()
	return activityOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(activity).Updates(updateInfo).Error
	})
}
