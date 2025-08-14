package database

import (
	"time"
)

type User struct {
	ID              uint   `gorm:"primarykey"`
	Username        string `gorm:"size:64;uniqueIndex;not null"`
	Email           string `gorm:"size:128;uniqueIndex;not null"`
	Cid             int    `gorm:"uniqueIndex;not null"`
	Password        string `gorm:"size:128;not null"`
	QQ              int    `gorm:"default:0"`
	Rating          int    `gorm:"default:0"`
	Permission      int64  `gorm:"default:0"`
	TotalPilotTime  int    `gorm:"default:0"`
	TotalAtcTime    int    `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FlightPlans     []FlightPlan    `gorm:"foreignKey:Cid;references:Cid"`
	OnlineHistories []History       `gorm:"foreignKey:Cid;references:Cid"`
	ActivityAtc     []ActivityATC   `gorm:"foreignKey:Cid;references:Cid"`
	ActivityPilot   []ActivityPilot `gorm:"foreignKey:Cid;references:Cid"`
}

type FlightPlan struct {
	ID               uint   `gorm:"primarykey"`
	Cid              int    `gorm:"index;not null"`
	Callsign         string `gorm:"size:16;uniqueIndex;not null"`
	FlightType       string `gorm:"size:4;not null"`
	AircraftType     string `gorm:"size:16;not null"`
	Tas              int    `gorm:"not null"`
	DepartureAirport string `gorm:"size:4;not null"`
	DepartureTime    int    `gorm:"not null"`
	AtcDepartureTime int    `gorm:"not null"`
	CruiseAltitude   string `gorm:"size:8;not null"`
	ArrivalAirport   string `gorm:"size:4;not null"`
	RouteTimeHour    string `gorm:"size:2;not null"`
	RouteTimeMinute  string `gorm:"size:2;not null"`
	FuelTimeHour     string `gorm:"size:2;not null"`
	FuelTimeMinute   string `gorm:"size:2;not null"`
	AlternateAirport string `gorm:"size:4;not null"`
	Remarks          string `gorm:"type:text;not null"`
	Route            string `gorm:"type:text;not null"`
	Locked           bool   `gorm:"default:0;not null"`
	FromWeb          bool   `gorm:"default:0;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type History struct {
	ID         uint      `gorm:"primarykey"`
	Cid        int       `gorm:"index;not null"`
	Callsign   string    `gorm:"size:16;index;not null"`
	StartTime  time.Time `gorm:"not null"`
	EndTime    time.Time `gorm:"not null"`
	OnlineTime int       `gorm:"default:0;not null"`
	IsAtc      bool      `gorm:"default:0;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Activity struct {
	ID               uint            `gorm:"primarykey"`
	Publisher        int             `gorm:"index;not null"`
	Title            string          `gorm:"size:128;not null"`
	ImageUrl         string          `gorm:"size:128;not null"`
	ActiveTime       time.Time       `gorm:"not null"`
	DepartureAirport string          `gorm:"size:4;not null"`
	ArrivalAirport   string          `gorm:"size:4;not null"`
	Route            string          `gorm:"size:128;not null"`
	Distance         int             `gorm:"default:0;not null"`
	Status           int             `gorm:"default:0;not null"`
	NOTAMS           string          `gorm:"type:text;not null"`
	Facilities       []ActivityATC   `gorm:"foreignKey:ActivityId;references:ID"`
	Pilots           []ActivityPilot `gorm:"foreignKey:ActivityId;references:ID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ActivityATC struct {
	ID         uint   `gorm:"primarykey"`
	ActivityId uint   `gorm:"index;not null"`
	Cid        uint   `gorm:"index;not null"`
	MinRating  int    `gorm:"default:2;not null"`
	Callsign   string `gorm:"size:16;not null"`
	Frequency  string `gorm:"size:4;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ActivityPilot struct {
	ID           uint   `gorm:"primarykey"`
	ActivityId   uint   `gorm:"index;not null"`
	Cid          int    `gorm:"index;not null"`
	Callsign     string `gorm:"size:16;not null"`
	AircraftType string `gorm:"size:8;not null"`
	Status       int    `gorm:"default:0;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
