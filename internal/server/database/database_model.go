package database

import (
	"time"
)

type User struct {
	ID              uint   `gorm:"primarykey"`
	Username        string `gorm:"size:64;uniqueIndex"`
	Email           string `gorm:"size:64;uniqueIndex"`
	Cid             int    `gorm:"uniqueIndex"`
	Password        string `gorm:"size:128"`
	QQ              int
	Rating          int
	Permission      int64
	TotalPilotTime  int
	TotalAtcTime    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FlightPlans     []FlightPlan    `gorm:"foreignKey:Cid;references:Cid"`
	OnlineHistories []History       `gorm:"foreignKey:Cid;references:Cid"`
	ActivityAtc     []ActivityATC   `gorm:"foreignKey:Cid;references:Cid"`
	ActivityPilot   []ActivityPilot `gorm:"foreignKey:Cid;references:Cid"`
}

type FlightPlan struct {
	ID               uint   `gorm:"primarykey"`
	Cid              int    `gorm:"index"`
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
}

type History struct {
	ID         uint   `gorm:"primarykey"`
	Cid        int    `gorm:"index"`
	Callsign   string `gorm:"size:16;index"`
	StartTime  time.Time
	EndTime    time.Time
	OnlineTime int
	IsAtc      bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Activity struct {
	ID               uint   `gorm:"primarykey"`
	Publisher        int    `gorm:"index"`
	Title            string `gorm:"size:128"`
	ImageUrl         string `gorm:"size:128"`
	ActiveTime       time.Time
	DepartureAirport string `gorm:"size:4"`
	ArrivalAirport   string `gorm:"size:4"`
	Route            string `gorm:"size:128"`
	Distance         int
	Status           int
	NOTAMS           string          `gorm:"type:text"`
	Facilities       []ActivityATC   `gorm:"foreignKey:ActivityId;references:ID"`
	Pilots           []ActivityPilot `gorm:"foreignKey:ActivityId;references:ID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ActivityATC struct {
	ID         uint `gorm:"primarykey"`
	ActivityId uint `gorm:"index"`
	Cid        uint `gorm:"index"`
	MinRating  int
	Callsign   string `gorm:"size:16"`
	Frequency  string `gorm:"size:4"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ActivityPilot struct {
	ID           uint   `gorm:"primarykey"`
	ActivityId   uint   `gorm:"index"`
	Cid          int    `gorm:"index"`
	Callsign     string `gorm:"size:16"`
	AircraftType string `gorm:"size:8"`
	Status       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
