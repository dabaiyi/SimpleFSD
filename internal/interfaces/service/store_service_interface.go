// Package service
package service

import (
	"fmt"
	c "github.com/half-nothing/simple-fsd/internal/config"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	ErrFilePathFail       = ApiStatus{"FILE_PATH_FAIL", "文件上传失败", ServerInternalError}
	ErrFileSaveFail       = ApiStatus{"FILE_SAVE_FAIL", "文件保存失败", ServerInternalError}
	ErrFileUploadFail     = ApiStatus{"FILE_UPLOAD_FAIL", "文件上传失败", ServerInternalError}
	ErrFileOverSize       = ApiStatus{"FILE_OVER_SIZE", "文件过大", BadRequest}
	ErrFileExtUnsupported = ApiStatus{"FILE_EXT_UNSUPPORTED", "不支持的文件类型", BadRequest}
	ErrFileNameIllegal    = ApiStatus{"FILE_NAME_ILLEGAL", "文件名不合法", BadRequest}
	SuccessUploadFIle     = ApiStatus{"UPLOAD_FILE", "文件上传成功", Ok}
)

type FileType int

const (
	IMAGES FileType = iota
	FILES
	UNKNOWN
)

// StoreInfo 文件存储信息
type StoreInfo struct {
	FileType      FileType                    // 文件类型 [FileType]
	FileLimit     *c.HttpServerStoreFileLimit // 该类型文件限制 [c.HttpServerStoreFileLimit]
	RootPath      string                      // 存储根目录
	FilePath      string                      // 文件存储路径
	RemotePath    string                      // 远程文件存储路径
	FileName      string                      // 文件名
	FileExt       string                      // 文件扩展名
	FileContent   *multipart.FileHeader       // 文件内容 [multipart.FileHeader]
	StoreInServer bool                        // 是否保存在本地
}

func NewStoreInfo(fileType FileType, fileLimit *c.HttpServerStoreFileLimit, file *multipart.FileHeader) *StoreInfo {
	return &StoreInfo{
		FileType:      fileType,
		FileLimit:     fileLimit,
		RootPath:      fileLimit.RootPath,
		FilePath:      "",
		FileName:      "",
		RemotePath:    "",
		FileExt:       filepath.Ext(file.Filename),
		FileContent:   file,
		StoreInServer: fileLimit.StoreInServer,
	}
}

func (fileType FileType) GenerateStoreInfo(fileLimit *c.HttpServerStoreFileLimit, file *multipart.FileHeader) (*StoreInfo, *ApiStatus) {
	if strings.Contains(file.Filename, string(filepath.Separator)) {
		return nil, &ErrFileNameIllegal
	}

	ext := filepath.Ext(file.Filename)

	if !slices.Contains(fileLimit.AllowedFileExt, ext) {
		return nil, &ErrFileExtUnsupported
	}

	if file.Size > fileLimit.MaxFileSize {
		return nil, &ErrFileOverSize
	}

	storeInfo := NewStoreInfo(fileType, fileLimit, file)

	storeInfo.FileName = filepath.Join(fileLimit.StorePrefix, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	storeInfo.FilePath = filepath.Join(fileLimit.RootPath, storeInfo.FileName)
	storeInfo.RemotePath = strings.Replace(storeInfo.FileName, "\\", "/", -1)

	return storeInfo, nil
}

type StoreServiceInterface interface {
	SaveImageFile(file *multipart.FileHeader) (*StoreInfo, *ApiStatus)
	SaveUploadImages(req *RequestUploadFile) *ApiResponse[ResponseUploadFile]
}

type RequestUploadFile struct {
	JwtHeader
	File *multipart.FileHeader
}

type ResponseUploadFile struct {
	FileSize   int64  `json:"file_size"`
	AccessPath string `json:"access_path"`
}
