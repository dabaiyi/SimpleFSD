// Package database
package database

import (
	"context"
	"time"
)

func NewHistory(cid int, callsign string, isAtc bool) *History {
	return &History{
		Cid:        cid,
		Callsign:   callsign,
		StartTime:  time.Now(),
		EndTime:    time.Now(),
		OnlineTime: 0,
		IsAtc:      isAtc,
	}
}

func (h *History) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(h).Error
	return err
}

func (h *History) End() error {
	h.EndTime = time.Now()
	h.OnlineTime = int(h.EndTime.Sub(h.StartTime).Seconds())
	return h.Save()
}
