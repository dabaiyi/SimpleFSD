// Package service
package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"github.com/labstack/echo/v4"
	"time"
)

type HttpCode int

const (
	Unsatisfied         HttpCode = 0
	Ok                  HttpCode = 200
	BadRequest          HttpCode = 400
	Unauthorized        HttpCode = 401
	PermissionDenied    HttpCode = 403
	NotFound            HttpCode = 404
	Conflict            HttpCode = 409
	ServerInternalError HttpCode = 500
)

func (hc HttpCode) Code() int {
	return int(hc)
}

type ApiStatus struct {
	StatusName  string
	Description string
	HttpCode    HttpCode
}

type ApiResponse[T any] struct {
	HttpCode int    `json:"-"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Data     *T     `json:"data"`
}

type Claims struct {
	Uid        uint   `json:"uid"`
	Cid        int    `json:"cid"`
	Username   string `json:"username"`
	Permission int64  `json:"permission"`
	Rating     int    `json:"rating"`
	FlushToken bool   `json:"flushToken"`
	config     *c.JWTConfig
	jwt.RegisteredClaims
}

type JwtHeader struct {
	Uid        uint
	Permission int64
}

func NewClaims(config *c.JWTConfig, user *operation.User, flushToken bool) *Claims {
	expiredDuration := config.ExpiresDuration
	if flushToken {
		expiredDuration += config.RefreshDuration
	}
	return &Claims{
		Uid:        user.ID,
		Cid:        user.Cid,
		Username:   user.Username,
		Permission: user.Permission,
		Rating:     user.Rating,
		FlushToken: flushToken,
		config:     config,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "FsdHttpServer",
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiredDuration)),
		},
	}
}

func (claim *Claims) GenerateKey() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	tokenString, _ := token.SignedString([]byte(claim.config.Secret))
	return tokenString
}

func (res *ApiResponse[T]) Response(ctx echo.Context) error {
	return ctx.JSON(res.HttpCode, res)
}

var (
	ErrIllegalParam          = ApiStatus{"PARAM_ERROR", "参数不正确", BadRequest}
	ErrLackParam             = ApiStatus{"PARAM_LACK_ERROR", "缺少参数", BadRequest}
	ErrNoPermission          = ApiStatus{"NO_PERMISSION", "无权这么做", PermissionDenied}
	ErrDatabaseFail          = ApiStatus{"DATABASE_ERROR", "服务器内部错误", ServerInternalError}
	ErrUserNotFound          = ApiStatus{"USER_NOT_FOUND", "指定用户不存在", NotFound}
	ErrActivityNotFound      = ApiStatus{"ACTIVITY_NOT_FOUND", "活动不存在", NotFound}
	ErrFacilityNotFound      = ApiStatus{"FACILITY_NOT_FOUND", "管制席位不存在", NotFound}
	ErrRegisterFail          = ApiStatus{"REGISTER_FAIL", "注册失败", ServerInternalError}
	ErrIdentifierTaken       = ApiStatus{"USER_EXISTS", "用户已存在", BadRequest}
	ErrMissingOrMalformedJwt = ApiStatus{"MISSING_OR_MALFORMED_JWT", "缺少JWT令牌或者令牌格式错误", BadRequest}
	ErrInvalidOrExpiredJwt   = ApiStatus{"INVALID_OR_EXPIRED_JWT", "无效或过期的JWT令牌", Unauthorized}
	ErrUnknown               = ApiStatus{"UNKNOWN_JWT_ERROR", "未知的JWT解析错误", ServerInternalError}
)

func NewErrorResponse(ctx echo.Context, codeStatus *ApiStatus) error {
	return NewApiResponse[any](codeStatus, Unsatisfied, nil).Response(ctx)
}

func NewApiResponse[T any](codeStatus *ApiStatus, httpCode HttpCode, data *T) *ApiResponse[T] {
	if httpCode == Unsatisfied {
		httpCode = codeStatus.HttpCode
	}
	if httpCode == Unsatisfied {
		httpCode = Ok
	}
	return &ApiResponse[T]{
		HttpCode: httpCode.Code(),
		Code:     codeStatus.StatusName,
		Message:  codeStatus.Description,
		Data:     data,
	}
}

// CallDBFuncAndCheckError 调用数据库操作函数并处理错误
func CallDBFuncAndCheckError[R any, T any](fc func() (*R, error)) (*R, *ApiResponse[T]) {
	result, err := fc()
	switch {
	case errors.Is(err, operation.ErrIdentifierCheck):
		return nil, NewApiResponse[T](&ErrRegisterFail, Unsatisfied, nil)
	case errors.Is(err, operation.ErrIdentifierTaken):
		return nil, NewApiResponse[T](&ErrIdentifierTaken, Unsatisfied, nil)
	case errors.Is(err, operation.ErrUserNotFound):
		return nil, NewApiResponse[T](&ErrUserNotFound, Unsatisfied, nil)
	case errors.Is(err, operation.ErrActivityNotFound):
		return nil, NewApiResponse[T](&ErrActivityNotFound, Unsatisfied, nil)
	case errors.Is(err, operation.ErrFacilityNotFound):
		return nil, NewApiResponse[T](&ErrFacilityNotFound, Unsatisfied, nil)
	case err != nil:
		c.ErrorF("Error in DB function: %v", err)
		return nil, NewApiResponse[T](&ErrDatabaseFail, Unsatisfied, nil)
	default:
		return result, nil
	}
}

// GetUsersAndCheckPermission 从数据库获取用户数据并检查权限
func GetUsersAndCheckPermission[T any](userOperation operation.UserOperationInterface, uid, targetUid uint, perm operation.Permission) (*operation.User, *operation.User, *ApiResponse[T]) {
	// 敏感操作获取实时数据
	user, res := CallDBFuncAndCheckError[operation.User, T](func() (*operation.User, error) { return userOperation.GetUserByUid(uid) })
	if res != nil {
		return nil, nil, res
	}
	permission := operation.Permission(user.Permission)
	if !permission.HasPermission(perm) {
		return nil, nil, NewApiResponse[T](&ErrNoPermission, Unsatisfied, nil)
	}
	targetUser, res := CallDBFuncAndCheckError[operation.User, T](func() (*operation.User, error) { return userOperation.GetUserByUid(targetUid) })
	if res != nil {
		return nil, nil, res
	}
	return user, targetUser, nil
}
