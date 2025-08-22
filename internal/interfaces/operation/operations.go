// Package operation
package operation

type DatabaseOperations struct {
	userOperation       UserOperationInterface
	flightPlanOperation FlightPlanOperationInterface
	historyOperation    HistoryOperationInterface
}

func NewDatabaseOperations(
	userOperation UserOperationInterface,
	flightPlanOperation FlightPlanOperationInterface,
	historyOperation HistoryOperationInterface,
) *DatabaseOperations {
	return &DatabaseOperations{
		userOperation:       userOperation,
		flightPlanOperation: flightPlanOperation,
		historyOperation:    historyOperation,
	}
}

func (db *DatabaseOperations) UserOperation() UserOperationInterface {
	return db.userOperation
}

func (db *DatabaseOperations) FlightPlanOperation() FlightPlanOperationInterface {
	return db.flightPlanOperation
}

func (db *DatabaseOperations) HistoryOperation() HistoryOperationInterface {
	return db.historyOperation
}
