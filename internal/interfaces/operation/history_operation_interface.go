// Package operation
package operation

type HistoryOperationInterface interface {
	NewHistory(cid int, callsign string, isAtc bool) (history *History)
	SaveHistory(history *History) (err error)
	EndRecordAndSaveHistory(history *History) (err error)
}
