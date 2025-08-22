package database

import (
	"context"
	. "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"gorm.io/gorm"
	"time"
)

type HistoryOperation struct {
	db *gorm.DB
}

func NewHistoryOperation(db *gorm.DB) *HistoryOperation {
	return &HistoryOperation{db: db}
}

func (historyOperation *HistoryOperation) NewHistory(cid int, callsign string, isAtc bool) (history *History) {
	return &History{
		Cid:        cid,
		Callsign:   callsign,
		StartTime:  time.Now(),
		EndTime:    time.Now(),
		OnlineTime: 0,
		IsAtc:      isAtc,
	}
}

func (historyOperation *HistoryOperation) SaveHistory(history *History) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	return database.WithContext(ctx).Save(history).Error
}

func (historyOperation *HistoryOperation) EndRecordAndSaveHistory(history *History) (err error) {
	history.EndTime = time.Now()
	history.OnlineTime = int(history.EndTime.Sub(history.StartTime).Seconds())
	return historyOperation.SaveHistory(history)
}
