// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/simple-fsd/internal/config"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/labstack/echo/v4"
)

type AuditLogControllerInterface interface {
	GetAuditLogs(ctx echo.Context) error
}

type AuditLogController struct {
	auditService AuditServiceInterface
}

func NewAuditLogController(auditService AuditServiceInterface) *AuditLogController {
	return &AuditLogController{auditService: auditService}
}

func (controller *AuditLogController) GetAuditLogs(ctx echo.Context) error {
	data := &RequestGetAuditLog{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.auditService.GetAuditLogPage(data).Response(ctx)
}
