// Package service
package service

import (
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
)

type AudioLogService struct {
	auditOperation operation.AuditLogOperationInterface
}

func NewAuditService(
	auditOperation operation.AuditLogOperationInterface,
) *AudioLogService {
	return &AudioLogService{
		auditOperation: auditOperation,
	}
}

var SuccessGetAuditLog = ApiStatus{StatusName: "GET_AUDIT_LOG", Description: "成功获取审计日志", HttpCode: Ok}

func (auditLogService *AudioLogService) GetAuditLogPage(req *RequestGetAuditLog) *ApiResponse[ResponseGetAuditLog] {
	if req.Page <= 0 || req.PageSize <= 0 {
		return NewApiResponse[ResponseGetAuditLog](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseGetAuditLog](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.AuditLogShow) {
		return NewApiResponse[ResponseGetAuditLog](&ErrNoPermission, Unsatisfied, nil)
	}
	auditLogs, total, err := auditLogService.auditOperation.GetAuditLogs(req.Page, req.PageSize)
	if err != nil {
		return NewApiResponse[ResponseGetAuditLog](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessGetAuditLog, Unsatisfied, &ResponseGetAuditLog{
		Items:    auditLogs,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	})
}
