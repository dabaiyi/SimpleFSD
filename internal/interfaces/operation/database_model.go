package operation

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID              uint             `gorm:"primarykey" json:"-"`
	Username        string           `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email           string           `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Cid             int              `gorm:"uniqueIndex;not null" json:"cid"`
	Password        string           `gorm:"size:128;not null" json:"-"`
	AvatarUrl       string           `gorm:"size:128;not null;default:''" json:"avatar_url"`
	QQ              int              `gorm:"default:0" json:"qq"`
	Rating          int              `gorm:"default:0" json:"rating"`
	Permission      int64            `gorm:"default:0" json:"permission"`
	TotalPilotTime  int              `gorm:"default:0" json:"total_pilot_time"`
	TotalAtcTime    int              `gorm:"default:0" json:"total_atc_time"`
	FlightPlans     []*FlightPlan    `gorm:"foreignKey:Cid;references:Cid" json:"-"`
	OnlineHistories []*History       `gorm:"foreignKey:Cid;references:Cid" json:"-"`
	ActivityAtc     []*ActivityATC   `gorm:"foreignKey:Cid;references:Cid" json:"-"`
	ActivityPilot   []*ActivityPilot `gorm:"foreignKey:Cid;references:Cid" json:"-"`
	CreatedAt       time.Time        `json:"-"`
	UpdatedAt       time.Time        `json:"-"`
}

type FlightPlan struct {
	ID               uint      `gorm:"primarykey" json:"-"`
	Cid              int       `gorm:"index;not null" json:"cid"`
	Callsign         string    `gorm:"size:16;uniqueIndex;not null" json:"callsign"`
	FlightType       string    `gorm:"size:4;not null" json:"flight_rules"`
	AircraftType     string    `gorm:"size:16;not null" json:"aircraft"`
	Tas              int       `gorm:"not null" json:"cruise_tas"`
	DepartureAirport string    `gorm:"size:4;not null" json:"departure"`
	DepartureTime    int       `gorm:"not null" json:"departure_time"`
	AtcDepartureTime int       `gorm:"not null" json:"-"`
	CruiseAltitude   string    `gorm:"size:8;not null" json:"altitude"`
	ArrivalAirport   string    `gorm:"size:4;not null" json:"arrival"`
	RouteTimeHour    string    `gorm:"size:2;not null" json:"route_time_hour"`
	RouteTimeMinute  string    `gorm:"size:2;not null" json:"route_time_minute"`
	FuelTimeHour     string    `gorm:"size:2;not null" json:"fuel_time_hour"`
	FuelTimeMinute   string    `gorm:"size:2;not null" json:"fuel_time_minute"`
	AlternateAirport string    `gorm:"size:4;not null" json:"alternate"`
	Remarks          string    `gorm:"type:text;not null" json:"remarks"`
	Route            string    `gorm:"type:text;not null" json:"route"`
	Locked           bool      `gorm:"default:0;not null" json:"-"`
	FromWeb          bool      `gorm:"default:0;not null" json:"-"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

type History struct {
	ID         uint      `gorm:"primarykey" json:"-"`
	Cid        int       `gorm:"index;not null" json:"-"`
	Callsign   string    `gorm:"size:16;index;not null" json:"callsign"`
	StartTime  time.Time `gorm:"not null" json:"start_time"`
	EndTime    time.Time `gorm:"not null" json:"end_time"`
	OnlineTime int       `gorm:"default:0;not null" json:"online_time"`
	IsAtc      bool      `gorm:"default:0;not null" json:"-"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type Activity struct {
	ID               uint                `gorm:"primarykey" json:"id"`
	Publisher        int                 `gorm:"index;not null" json:"publisher"`
	Title            string              `gorm:"size:128;not null" json:"title"`
	ImageUrl         string              `gorm:"size:128;not null" json:"image_url"`
	ActiveTime       time.Time           `gorm:"not null" json:"active_time"`
	DepartureAirport string              `gorm:"size:64;not null" json:"departure_airport"`
	ArrivalAirport   string              `gorm:"size:64;not null" json:"arrival_airport"`
	Route            string              `gorm:"size:128;not null" json:"route"`
	Distance         int                 `gorm:"default:0;not null" json:"distance"`
	Status           int                 `gorm:"default:0;not null" json:"status"`
	NOTAMS           string              `gorm:"type:text;not null" json:"NOTAMS"`
	Facilities       []*ActivityFacility `gorm:"foreignKey:ActivityId;references:ID" json:"facilities"`
	Controllers      []*ActivityATC      `gorm:"foreignKey:ActivityId;references:ID" json:"controllers"`
	Pilots           []*ActivityPilot    `gorm:"foreignKey:ActivityId;references:ID" json:"pilots"`
	CreatedAt        time.Time           `json:"-"`
	UpdatedAt        time.Time           `json:"-"`
	DeletedAt        gorm.DeletedAt      `json:"-"`
}

func (facility *Activity) Equal(other *Activity) bool {
	return facility.ID == other.ID && facility.Publisher == other.Publisher && facility.Title == other.Title &&
		facility.ImageUrl == other.ImageUrl && facility.ActiveTime == other.ActiveTime &&
		facility.DepartureAirport == other.DepartureAirport && facility.ArrivalAirport == other.ArrivalAirport &&
		facility.Route == other.Route && facility.Distance == other.Distance && facility.Status == other.Status &&
		facility.NOTAMS == other.NOTAMS
}

func (facility *Activity) Diff(other *Activity) map[string]interface{} {
	result := make(map[string]interface{})
	if facility.Publisher != 0 && facility.Publisher != other.Publisher {
		other.Publisher = facility.Publisher
		result["publisher"] = facility.Publisher
	}
	if facility.Title != "" && facility.Title != other.Title {
		other.Title = facility.Title
		result["title"] = facility.Title
	}
	if facility.ImageUrl != "" && facility.ImageUrl != other.ImageUrl {
		other.ImageUrl = facility.ImageUrl
		result["image_url"] = facility.ImageUrl
	}
	if facility.ActiveTime != other.ActiveTime {
		other.ActiveTime = facility.ActiveTime
		result["active_time"] = facility.ActiveTime
	}
	if facility.DepartureAirport != "" && facility.DepartureAirport != other.DepartureAirport {
		other.DepartureAirport = facility.DepartureAirport
		result["departure_airport"] = facility.DepartureAirport
	}
	if facility.ArrivalAirport != "" && facility.ArrivalAirport != other.ArrivalAirport {
		other.ArrivalAirport = facility.ArrivalAirport
		result["arrival_airport"] = facility.ArrivalAirport
	}
	if facility.Route != "" && facility.Route != other.Route {
		other.Route = facility.Route
		result["route"] = facility.Route
	}
	if facility.Distance != 0 && facility.Distance != other.Distance {
		other.Distance = facility.Distance
		result["distance"] = facility.Distance
	}
	if facility.Status != other.Status {
		other.Status = facility.Status
		result["status"] = facility.Status
	}
	if facility.NOTAMS != other.NOTAMS {
		other.NOTAMS = facility.NOTAMS
		result["NOTAMS"] = facility.NOTAMS
	}
	return result
}

type ActivityFacility struct {
	ID         uint         `gorm:"primarykey" json:"id"`
	ActivityId uint         `gorm:"index;not null" json:"activity_id"`
	MinRating  int          `gorm:"default:2;not null" json:"min_rating"`
	Callsign   string       `gorm:"size:16;not null" json:"callsign"`
	Frequency  string       `gorm:"size:8;not null" json:"frequency"`
	Controller *ActivityATC `gorm:"foreignKey:FacilityId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt  time.Time    `json:"-"`
	UpdatedAt  time.Time    `json:"-"`
}

func (facility *ActivityFacility) Equal(other *ActivityFacility) bool {
	return facility.ID == other.ID && facility.ActivityId == other.ActivityId && facility.MinRating == other.MinRating &&
		facility.Callsign == other.Callsign && facility.Frequency == other.Frequency
}

func (facility *ActivityFacility) Diff(other *ActivityFacility) map[string]interface{} {
	result := make(map[string]interface{})
	if facility.MinRating != other.MinRating {
		result["min_rating"] = facility.MinRating
	}
	if facility.Callsign != other.Callsign {
		result["callsign"] = facility.Callsign
	}
	if facility.Frequency != other.Frequency {
		result["frequency"] = facility.Frequency
	}
	return result
}

type ActivityATC struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ActivityId uint      `gorm:"uniqueIndex:activityController;not null" json:"activity_id"`
	FacilityId uint      `gorm:"uniqueIndex:activityController;not null" json:"facility_id"`
	Cid        int       `gorm:"index;not null" json:"cid"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type ActivityPilot struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	ActivityId   uint      `gorm:"uniqueIndex:activityPilot;not null" json:"activity_id"`
	Cid          int       `gorm:"uniqueIndex:activityPilot;not null" json:"cid"`
	Callsign     string    `gorm:"size:16;not null" json:"callsign"`
	AircraftType string    `gorm:"size:8;not null" json:"aircraft_type"`
	Status       int       `gorm:"default:0;not null" json:"status"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type AuditLog struct {
	ID            uint          `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time     `gorm:"not null" json:"time"`
	EventType     string        `gorm:"index:eventType;not null" json:"event_type"`
	Subject       int           `gorm:"index:Subject;not null" json:"subject"`
	Object        string        `gorm:"index:Object;not null" json:"object"`
	Ip            string        `gorm:"not null" json:"ip"`
	UserAgent     string        `gorm:"not null" json:"user_agent"`
	ChangeDetails *ChangeDetail `gorm:"type:text;serializer:json" json:"change_details"`
}

type ChangeDetail struct {
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}
