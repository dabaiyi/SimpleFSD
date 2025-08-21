// Package operation
package database

import (
	"context"
	database2 "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"time"
)

func NewHistory(cid int, callsign string, isAtc bool) *database2.History {
	return &database2.History{
		Cid:        cid,
		Callsign:   callsign,
		StartTime:  time.Now(),
		EndTime:    time.Now(),
		OnlineTime: 0,
		IsAtc:      isAtc,
	}
}

func (h *database2.History) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(h).Error
	return err
}

func (h *database2.History) End() error {
	h.EndTime = time.Now()
	h.OnlineTime = int(h.EndTime.Sub(h.StartTime).Seconds())
	return h.Save()
}
