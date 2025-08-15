// Package fsd
package fsd

import (
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/utils"
)

type FacilityModel struct {
	Id        int    `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
}

type Facility byte

const (
	Pilot Facility = 1 << iota
	OBS
	DEL
	GND
	TWR
	APP
	CTR
	FSS
)

var Facilities = []FacilityModel{
	{0, "Pilot", "Pilot"},
	{1, "OBS", "Observer"},
	{2, "DEL", "Clearance Delivery"},
	{3, "GND", "Ground"},
	{4, "TWR", "Tower"},
	{5, "APP", "Approach/Departure"},
	{6, "CTR", "Enroute"},
	{7, "FSS", "Flight Service Station"},
}

func (f Facility) String() string {
	return Facilities[f].ShortName
}

func (f Facility) Index() int {
	return int(f)
}

func (f Facility) CheckFacility(facility Facility) bool {
	return f&facility != 0
}

func (r Rating) CheckRatingFacility(facility Facility) bool {
	return RatingFacilityMap[r].CheckFacility(facility)
}

func SyncRatingConfig() error {
	config, _ := c.GetConfig()
	if len(config.Rating) == 0 {
		return nil
	}
	for rating, facility := range config.Rating {
		r := utils.StrToInt(rating, -2)
		if r <= -2 || r > 12 {
			return fmt.Errorf("illegal permission value %s", rating)
		}
		RatingFacilityMap[Rating(r)] = Facility(facility)
	}
	return nil
}
