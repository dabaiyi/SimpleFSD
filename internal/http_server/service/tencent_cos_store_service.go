// Package service
package service

import (
	"context"
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	. "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	. "github.com/half-nothing/fsd-server/internal/interfaces/service"
	"github.com/tencentyun/cos-go-sdk-v5"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type TencentCosStoreService struct {
	localStore StoreServiceInterface
	config     *c.HttpServerStore
	endpoint   *url.URL
	client     *cos.Client
}

func NewTencentCosStoreService(localStore StoreServiceInterface, config *c.HttpServerStore) *TencentCosStoreService {
	service := &TencentCosStoreService{localStore: localStore, config: config}
	bucketUrl, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, strings.ToLower(config.Region)))
	serviceUrl, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", strings.ToLower(config.Region)))
	baseUrl := &cos.BaseURL{BucketURL: bucketUrl, ServiceURL: serviceUrl}
	service.client = cos.NewClient(baseUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessId,
			SecretKey: config.AccessKey,
		},
	})
	if config.CdnDomain != "" {
		service.endpoint, _ = url.Parse(config.CdnDomain)
	} else {
		service.endpoint = service.client.BaseURL.BucketURL
	}
	return service
}

func (store *TencentCosStoreService) SaveImageFile(file *multipart.FileHeader) (*StoreInfo, *ApiStatus) {
	storeInfo, res := store.localStore.SaveImageFile(file)
	if res != nil {
		return nil, res
	}

	storeInfo.RemotePath = strings.Replace(filepath.Join(store.config.RemoteStorePath, storeInfo.FileName), "\\", "/", -1)

	reader, err := file.Open()
	if err != nil {
		c.ErrorF("Failed to open form file: %v", err)
		return nil, &ErrFileUploadFail
	}

	_, err = store.client.Object.Put(context.Background(), storeInfo.RemotePath, reader, nil)
	if err != nil {
		c.ErrorF("Failed to upload image to remote storage: %v", err)
		return nil, &ErrFileUploadFail
	}
	return storeInfo, nil
}

func (store *TencentCosStoreService) SaveUploadImages(req *RequestUploadFile) *ApiResponse[ResponseUploadFile] {
	if req.Permission <= 0 {
		return NewApiResponse[ResponseUploadFile](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := Permission(req.Permission)
	if !permission.HasPermission(ActivityPublish) {
		return NewApiResponse[ResponseUploadFile](&ErrNoPermission, PermissionDenied, nil)
	}
	storeInfo, res := store.SaveImageFile(req.File)
	if res != nil {
		return NewApiResponse[ResponseUploadFile](res, Unsatisfied, nil)
	}
	accessUrl, err := url.JoinPath(store.endpoint.String(), storeInfo.RemotePath)
	if err != nil {
		return NewApiResponse[ResponseUploadFile](&ErrFilePathFail, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessUploadFIle, Unsatisfied, &ResponseUploadFile{
		FileSize:   req.File.Size,
		AccessPath: accessUrl,
	})
}
