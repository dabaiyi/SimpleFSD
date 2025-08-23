// Package service
package service

import (
	. "github.com/half-nothing/simple-fsd/internal/interfaces/fsd"
)

type ServerServiceInterface interface {
	GetServerConfig() *ApiResponse[ResponseGetServerConfig]
}

type ServerLimits struct {
	UsernameLengthMin int  `json:"username_length_min"`
	UsernameLengthMax int  `json:"username_length_max"`
	PasswordLengthMin int  `json:"password_length_min"`
	PasswordLengthMax int  `json:"password_length_max"`
	EmailLengthMin    int  `json:"email_length_min"`
	EmailLengthMax    int  `json:"email_length_max"`
	CidMin            int  `json:"cid_min"`
	CidMax            int  `json:"cid_max"`
	SimulatorServer   bool `json:"simulator_server"`
}

type ServerConfig struct {
	Limits     *ServerLimits    `json:"limits"`
	Facilities *[]FacilityModel `json:"facilities"`
	Ratings    *[]RatingModel   `json:"ratings"`
}

type ResponseGetServerConfig struct {
	*ServerConfig
}
