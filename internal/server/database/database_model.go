package database

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	Username  string `gorm:"size:64;uniqueIndex:user_ident_index"`
	Email     string `gorm:"size:64;uniqueIndex:user_ident_index"`
	Cid       int    `gorm:"uniqueIndex:user_ident_index"`
	Password  string `gorm:"size:64"`
	QQ        int
	Rating    int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type FlightPlan struct {
	ID               uint   `gorm:"primarykey"`
	Callsign         string `gorm:"size:16;uniqueIndex"`
	FlightType       string `gorm:"size:4"`
	AircraftType     string `gorm:"size:16"`
	Tas              int
	DepartureAirport string `gorm:"size:4"`
	DepartureTime    int
	AtcDepartureTime int
	CruiseAltitude   string `gorm:"size:8"`
	ArrivalAirport   string `gorm:"size:4"`
	RouteTime        int
	FuelTime         int
	AlternateAirport string `gorm:"size:4"`
	Remarks          string `gorm:"type:text"`
	Route            string `gorm:"type:text"`
	Locked           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}
