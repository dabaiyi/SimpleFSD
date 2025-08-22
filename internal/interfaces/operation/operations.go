// Package operation
package operation

type DatabaseOperations struct {
	userOperation       UserOperationInterface
	flightPlanOperation FlightPlanOperationInterface
	historyOperation    HistoryOperationInterface
	activityOperation   ActivityOperationInterface
}

func NewDatabaseOperations(
	userOperation UserOperationInterface,
	flightPlanOperation FlightPlanOperationInterface,
	historyOperation HistoryOperationInterface,
	activityOperation ActivityOperationInterface,
) *DatabaseOperations {
	return &DatabaseOperations{
		userOperation:       userOperation,
		flightPlanOperation: flightPlanOperation,
		historyOperation:    historyOperation,
		activityOperation:   activityOperation,
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

func (db *DatabaseOperations) ActivityOperation() ActivityOperationInterface {
	return db.activityOperation
}
