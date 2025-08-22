// Package operation
package operation

// HistoryOperationInterface 联飞记录操作接口定义
type HistoryOperationInterface interface {
	// NewHistory 创建新联飞记录
	NewHistory(cid int, callsign string, isAtc bool) (history *History)
	// SaveHistory 保存联飞记录到数据库, 当err为nil时保存成功
	SaveHistory(history *History) (err error)
	// EndRecordAndSaveHistory 结束联飞记录并保存到数据库, 当err为nil时保存成功
	EndRecordAndSaveHistory(history *History) (err error)
}
