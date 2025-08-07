package database

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID          uint   `gorm:"primarykey"`
	Username    string `gorm:"size:64;uniqueIndex:user_ident_index"`
	Email       string `gorm:"size:64;uniqueIndex:user_ident_index"`
	Cid         int    `gorm:"uniqueIndex"`
	Password    string `gorm:"size:64"`
	QQ          int
	Rating      int
	FlightPlans FlightPlan `gorm:"foreignKey:Cid;references:Cid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type FlightPlan struct {
	ID               uint   `gorm:"primarykey"`
	Cid              int    `gorm:"uniqueIndex"`
	Callsign         string `gorm:"size:16;uniqueIndex"`
	FlightType       string `gorm:"size:4"`
	AircraftType     string `gorm:"size:16"`
	Tas              int
	DepartureAirport string `gorm:"size:4"`
	DepartureTime    int
	AtcDepartureTime int
	CruiseAltitude   string `gorm:"size:8"`
	ArrivalAirport   string `gorm:"size:4"`
	RouteTimeHour    string `gorm:"size:2"`
	RouteTimeMinute  string `gorm:"size:2"`
	FuelTimeHour     string `gorm:"size:2"`
	FuelTimeMinute   string `gorm:"size:2"`
	AlternateAirport string `gorm:"size:4"`
	Remarks          string `gorm:"type:text"`
	Route            string `gorm:"type:text"`
	Locked           bool
	FromWeb          bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}
