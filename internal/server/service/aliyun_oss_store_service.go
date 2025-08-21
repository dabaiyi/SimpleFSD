// Package service
package service

import (
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	c "github.com/half-nothing/fsd-server/internal/config"
	. "github.com/half-nothing/fsd-server/internal/server/defination"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
)

type ALiYunOssStoreService struct {
	localStore StoreServiceInterface
	config     *c.HttpServerStore
	endpoint   *url.URL
	client     *oss.Client
}

func NewALiYunOssStoreService(localStore StoreServiceInterface, config *c.HttpServerStore) *ALiYunOssStoreService {
	service := &ALiYunOssStoreService{localStore: localStore, config: config}
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessId, config.AccessKey)).
		WithRegion(config.Region).
		WithUseInternalEndpoint(config.UseInternalUrl)
	service.client = oss.NewClient(cfg)
	if config.CdnDomain != "" {
		service.endpoint, _ = url.Parse(config.CdnDomain)
	} else {
		service.endpoint, _ = url.Parse(strings.Replace(*cfg.Endpoint, "-internal", "", 1))
	}
	return service
}

func (store *ALiYunOssStoreService) SaveImageFile(file *multipart.FileHeader) (*StoreInfo, *ApiStatus) {
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

	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(store.config.Bucket),
		Key:          oss.Ptr(storeInfo.RemotePath),
		StorageClass: oss.StorageClassStandard,
		Body:         reader,
	}

	_, err = store.client.PutObject(context.TODO(), putRequest)

	if err != nil {
		c.ErrorF("Failed to upload image to remote storage: %v", err)
		return nil, &ErrFileUploadFail
	}
	return storeInfo, nil
}

func (store *ALiYunOssStoreService) SaveUploadImages(req *RequestUploadFile) *ApiResponse[ResponseUploadFile] {
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
