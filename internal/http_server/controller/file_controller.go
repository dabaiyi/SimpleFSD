// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/labstack/echo/v4"
)

type FileControllerInterface interface {
	UploadImages(ctx echo.Context) error
}

type FileController struct {
	storeService StoreServiceInterface
}

func NewFileController(storeService StoreServiceInterface) *FileController {
	return &FileController{storeService: storeService}
}

func (controller *FileController) UploadImages(ctx echo.Context) error {
	if file, err := ctx.FormFile("file"); err != nil {
		return NewErrorResponse(ctx, &ErrLackParam)
	} else {
		data := &RequestUploadFile{File: file}
		token := ctx.Get("user").(*jwt.Token)
		claim := token.Claims.(*Claims)
		data.Uid = claim.Uid
		data.Permission = claim.Permission
		return controller.storeService.SaveUploadImages(data).Response(ctx)
	}
}
