// Package service
package service

import (
	"fmt"
	c "github.com/half-nothing/simple-fsd/internal/config"
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type LocalStoreService struct {
	config *c.HttpServerStore
}

func NewLocalStoreService(config *c.HttpServerStore) *LocalStoreService {
	return &LocalStoreService{config}
}

func (store *LocalStoreService) GetFileStoreInfo(limit *c.HttpServerStoreFileLimit, file *multipart.FileHeader) (string, string, *ApiStatus) {
	if strings.Contains(file.Filename, string(filepath.Separator)) {
		return "", "", &ErrFileNameIllegal
	}

	ext := filepath.Ext(file.Filename)

	if !slices.Contains(limit.AllowedFileExt, ext) {
		return "", "", &ErrFileExtUnsupported
	}

	if file.Size > limit.MaxFileSize {
		return "", "", &ErrFileOverSize
	}

	newFilename := filepath.Join(limit.StorePrefix, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	dstPath := filepath.Join(store.config.LocalStorePath, newFilename)
	return dstPath, newFilename, nil
}

func (store *LocalStoreService) SaveImageFile(file *multipart.FileHeader) (*StoreInfo, *ApiStatus) {
	storeInfo, res := IMAGES.GenerateStoreInfo(store.config.FileLimit.ImageLimit, file)
	if res != nil {
		return nil, res
	}
	if !storeInfo.StoreInServer {
		return storeInfo, nil
	}
	src, err := file.Open()
	defer func(src multipart.File) {
		_ = src.Close()
	}(src)
	if err != nil {
		c.ErrorF("open file error: %v", err)
		return nil, &ErrFileSaveFail
	}
	dst, err := os.Create(storeInfo.FilePath)
	defer func(dst *os.File) {
		_ = dst.Close()
	}(dst)
	if err != nil {
		c.ErrorF("create file error: %v", err)
		return nil, &ErrFileSaveFail
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		c.ErrorF("copy file error: %v", err)
		return nil, &ErrFileSaveFail
	}
	return storeInfo, nil
}

func (store *LocalStoreService) DeleteImageFile(file string) (*StoreInfo, error) {
	storeInfo := NewStoreInfo(IMAGES, store.config.FileLimit.ImageLimit, nil)

	storeInfo.FileName = filepath.Join(store.config.FileLimit.ImageLimit.StorePrefix, filepath.Base(file))
	storeInfo.FilePath = filepath.Join(store.config.FileLimit.ImageLimit.RootPath, storeInfo.FileName)
	storeInfo.RemotePath = strings.Replace(storeInfo.FileName, "\\", "/", -1)

	if !storeInfo.StoreInServer {
		return storeInfo, nil
	}

	if err := os.Remove(storeInfo.FilePath); err != nil {
		c.ErrorF("remove file error: %v", err)
		return nil, err
	}
	return storeInfo, nil
}

func (store *LocalStoreService) SaveUploadImages(req *RequestUploadFile) *ApiResponse[ResponseUploadFile] {
	if req.Permission <= 0 {
		return NewApiResponse[ResponseUploadFile](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ActivityPublish) {
		return NewApiResponse[ResponseUploadFile](&ErrNoPermission, PermissionDenied, nil)
	}
	storeInfo, res := store.SaveImageFile(req.File)
	if res != nil {
		return NewApiResponse[ResponseUploadFile](res, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessUploadFile, Unsatisfied, &ResponseUploadFile{
		FileSize:   req.File.Size,
		AccessPath: storeInfo.RemotePath,
	})
}
