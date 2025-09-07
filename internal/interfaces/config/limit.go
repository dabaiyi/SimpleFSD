// Package config
package config

import (
	"errors"
	"github.com/half-nothing/simple-fsd/internal/interfaces/log"
	"time"
)

type HttpServerLimit struct {
	RateLimit         int           `json:"rate_limit"`
	RateLimitWindow   string        `json:"rate_limit_window"`
	RateLimitDuration time.Duration `json:"-"`
	UsernameLengthMin int           `json:"username_length_min"`
	UsernameLengthMax int           `json:"username_length_max"`
	EmailLengthMin    int           `json:"email_length_min"`
	EmailLengthMax    int           `json:"email_length_max"`
	PasswordLengthMin int           `json:"password_length_min"`
	PasswordLengthMax int           `json:"password_length_max"`
	CidMin            int           `json:"cid_min"`
	CidMax            int           `json:"cid_max"`
}

func defaultHttpServerLimit() *HttpServerLimit {
	return &HttpServerLimit{
		RateLimit:         15,
		RateLimitWindow:   "1m",
		UsernameLengthMin: 4,
		UsernameLengthMax: 16,
		EmailLengthMin:    4,
		EmailLengthMax:    64,
		PasswordLengthMin: 6,
		PasswordLengthMax: 64,
		CidMin:            1,
		CidMax:            9999,
	}
}

func (config *HttpServerLimit) checkValid(_ log.LoggerInterface) *ValidResult {
	if duration, err := time.ParseDuration(config.RateLimitWindow); err != nil {
		return ValidFailWith(errors.New("invalid json field http_server.rate_limit_window, %v"), err)
	} else {
		config.RateLimitDuration = duration
	}

	if config.UsernameLengthMin <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.username_length_min, value must larger than 0"))
	}
	if config.UsernameLengthMin > 64 {
		return ValidFail(errors.New("invalid json field http_server.limits.username_length_min, value must less than 64"))
	}
	if config.UsernameLengthMax <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.username_length_max, value must larger than 0"))
	}
	if config.UsernameLengthMax > 64 {
		return ValidFail(errors.New("invalid json field http_server.limits.username_length_max, value must less than 64"))
	}
	if config.UsernameLengthMin >= config.UsernameLengthMax {
		return ValidFail(errors.New("invalid json field http_server.limits.username_length_min, value must less than http_server.limits.username_length_max"))
	}

	if config.EmailLengthMin <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.email_length_min, value must larger than 0"))
	}
	if config.EmailLengthMin > 128 {
		return ValidFail(errors.New("invalid json field http_server.limits.email_length_min, value must less than 128"))
	}
	if config.EmailLengthMax <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.email_length_max, value must larger than 0"))
	}
	if config.EmailLengthMax > 128 {
		return ValidFail(errors.New("invalid json field http_server.limits.email_length_max, value must less than 128"))
	}
	if config.EmailLengthMin >= config.EmailLengthMax {
		return ValidFail(errors.New("invalid json field http_server.limits.email_length_min, value must less than http_server.limits.email_length_max"))
	}

	if config.PasswordLengthMin <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.password_length_min, value must larger than 0"))
	}
	if config.PasswordLengthMin > 128 {
		return ValidFail(errors.New("invalid json field http_server.limits.password_length_min, value must less than 128"))
	}
	if config.PasswordLengthMax <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.password_length_max, value must larger than 0"))
	}
	if config.PasswordLengthMax > 128 {
		return ValidFail(errors.New("invalid json field http_server.limits.password_length_max, value must less than 128"))
	}
	if config.PasswordLengthMin >= config.PasswordLengthMax {
		return ValidFail(errors.New("invalid json field http_server.limits.password_length_min, value must less than http_server.limits.password_length_max"))
	}

	if config.CidMin <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.cid_min, value must larger than 0"))
	}
	if config.CidMax <= 0 {
		return ValidFail(errors.New("invalid json field http_server.limits.cid_max, value must larger than 0"))
	}
	if config.CidMin >= config.CidMax {
		return ValidFail(errors.New("invalid json field http_server.limits.cid_min, value must less than http_server.limits.cid_max"))
	}

	return ValidPass()
}
