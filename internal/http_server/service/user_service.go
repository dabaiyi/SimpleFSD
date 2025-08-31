// Package service
package service

import (
	"errors"
	"fmt"
	c "github.com/half-nothing/simple-fsd/internal/config"
	"github.com/half-nothing/simple-fsd/internal/interfaces/fsd"
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/half-nothing/simple-fsd/internal/utils"
	"strings"
	"time"
)

type UserService struct {
	emailService     EmailServiceInterface
	config           *c.HttpServerConfig
	userOperation    operation.UserOperationInterface
	historyOperation operation.HistoryOperationInterface
	storeService     StoreServiceInterface
}

func NewUserService(
	emailService EmailServiceInterface,
	config *c.HttpServerConfig,
	userOperation operation.UserOperationInterface,
	historyOperation operation.HistoryOperationInterface,
	storeService StoreServiceInterface,
) *UserService {
	return &UserService{
		emailService:     emailService,
		config:           config,
		userOperation:    userOperation,
		historyOperation: historyOperation,
		storeService:     storeService,
	}
}

var (
	ErrEmailNotFound    = ApiStatus{StatusName: "EMAIL_CODE_NOT_FOUND", Description: "未向该邮箱发送验证码", HttpCode: BadRequest}
	ErrCidNotMatch      = ApiStatus{StatusName: "CID_NOT_MATCH", Description: "注册cid与验证码发送时的cid不一致", HttpCode: BadRequest}
	ErrEmailExpired     = ApiStatus{StatusName: "EMAIL_CODE_EXPIRED", Description: "验证码已过期", HttpCode: BadRequest}
	ErrEmailCodeInvalid = ApiStatus{StatusName: "EMAIL_CODE_INVALID", Description: "邮箱验证码错误", HttpCode: BadRequest}
	SuccessRegister     = ApiStatus{StatusName: "REGISTER_SUCCESS", Description: "注册成功", HttpCode: Ok}
)

func (userService *UserService) verifyEmailCode(email string, emailCode, cid int) *ApiStatus {
	err := userService.emailService.VerifyCode(email, emailCode, cid)
	switch {
	case errors.Is(err, ErrEmailCodeNotFound):
		return &ErrEmailNotFound
	case errors.Is(err, ErrEmailCodeExpired):
		return &ErrEmailExpired
	case errors.Is(err, ErrInvalidEmailCode):
		return &ErrEmailCodeInvalid
	case errors.Is(err, ErrCidMismatch):
		return &ErrCidNotMatch
	default:
		return nil
	}
}

func (userService *UserService) UserRegister(req *RequestUserRegister) *ApiResponse[ResponseUserRegister] {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.Cid <= 0 || req.EmailCode < 1e5 {
		return NewApiResponse[ResponseUserRegister](&ErrIllegalParam, Unsatisfied, nil)
	}
	if res := userService.verifyEmailCode(req.Email, req.EmailCode, req.Cid); res != nil {
		return NewApiResponse[ResponseUserRegister](res, Unsatisfied, nil)
	}
	if err := usernameValidator.CheckString(req.Username); err != nil {
		return NewApiResponse[ResponseUserRegister](err, Unsatisfied, nil)
	}
	if err := emailValidator.CheckString(req.Email); err != nil {
		return NewApiResponse[ResponseUserRegister](err, Unsatisfied, nil)
	}
	if err := passwordValidator.CheckString(req.Password); err != nil {
		return NewApiResponse[ResponseUserRegister](err, Unsatisfied, nil)
	}
	if err := cidValidator.CheckInt(req.Cid); err != nil {
		return NewApiResponse[ResponseUserRegister](err, Unsatisfied, nil)
	}
	user, err := userService.userOperation.NewUser(req.Username, req.Email, req.Cid, req.Password)
	if err != nil {
		return NewApiResponse[ResponseUserRegister](&ErrRegisterFail, Unsatisfied, nil)
	}
	if _, res := CallDBFuncAndCheckError[interface{}, ResponseUserRegister](func() (*interface{}, error) {
		return nil, userService.userOperation.AddUser(user)
	}); res != nil {
		return res
	}
	token := NewClaims(userService.config.JWT, user, false)
	flushToken := NewClaims(userService.config.JWT, user, true)
	return NewApiResponse(&SuccessRegister, Unsatisfied, &ResponseUserRegister{
		User:       user,
		Token:      token.GenerateKey(),
		FlushToken: flushToken.GenerateKey(),
	})
}

var (
	ErrUsernameOrPassword = ApiStatus{StatusName: "WRONG_USERNAME_OR_PASSWORD", Description: "用户名或密码错误", HttpCode: BadRequest}
	SuccessLogin          = ApiStatus{StatusName: "LOGIN_SUCCESS", Description: "登陆成功", HttpCode: Ok}
)

func (userService *UserService) UserLogin(req *RequestUserLogin) *ApiResponse[ResponseUserLogin] {
	if req.Username == "" || req.Password == "" {
		return NewApiResponse[ResponseUserLogin](&ErrIllegalParam, Unsatisfied, nil)
	}
	userId := operation.GetUserId(req.Username)

	user, res := CallDBFuncAndCheckError[operation.User, ResponseUserLogin](func() (*operation.User, error) {
		return userId.GetUser(userService.userOperation)
	})
	if res != nil {
		return res
	}

	if pass := userService.userOperation.VerifyUserPassword(user, req.Password); pass {
		token := NewClaims(userService.config.JWT, user, false)
		flushToken := NewClaims(userService.config.JWT, user, true)
		return NewApiResponse(&SuccessLogin, Unsatisfied, &ResponseUserLogin{
			User:       user,
			Token:      token.GenerateKey(),
			FlushToken: flushToken.GenerateKey(),
		})
	} else {
		return NewApiResponse[ResponseUserLogin](&ErrUsernameOrPassword, Unsatisfied, nil)
	}
}

var (
	NameNotAvailability = ApiStatus{StatusName: "INFO_NOT_AVAILABILITY", Description: "用户信息不可用", HttpCode: Ok}
	NameAvailability    = ApiStatus{StatusName: "INFO_AVAILABILITY", Description: "用户信息可用", HttpCode: Ok}
)

func (userService *UserService) CheckAvailability(req *RequestUserAvailability) *ApiResponse[ResponseUserAvailability] {
	if req.Username == "" && req.Email == "" && req.Cid == "" {
		return NewApiResponse[ResponseUserAvailability](&ErrIllegalParam, Unsatisfied, nil)
	}
	exist, _ := userService.userOperation.IsUserIdentifierTaken(nil, utils.StrToInt(req.Cid, 0), req.Username, req.Email)
	data := ResponseUserAvailability(!exist)
	if exist {
		return NewApiResponse(&NameNotAvailability, Unsatisfied, &data)
	}
	return NewApiResponse(&NameAvailability, Unsatisfied, &data)
}

var (
	SuccessGetCurrentProfile = ApiStatus{StatusName: "GET_CURRENT_PROFILE_SUCCESS", Description: "获取当前用户信息成功", HttpCode: Ok}
)

func (userService *UserService) GetCurrentProfile(req *RequestUserCurrentProfile) *ApiResponse[ResponseUserCurrentProfile] {
	if user, err := userService.userOperation.GetUserByUid(req.Uid); errors.Is(err, operation.ErrUserNotFound) {
		return NewApiResponse[ResponseUserCurrentProfile](&ErrUserNotFound, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[ResponseUserCurrentProfile](&ErrDatabaseFail, Unsatisfied, nil)
	} else {
		return NewApiResponse(&SuccessGetCurrentProfile, Unsatisfied, (*ResponseUserCurrentProfile)(user))
	}
}

var (
	ErrOriginPasswordRequired = ApiStatus{StatusName: "ORIGIN_PASSWORD_REQUIRED", Description: "请输入原始密码", HttpCode: BadRequest}
	ErrNewPasswordRequired    = ApiStatus{StatusName: "NEW_PASSWORD_REQUIRED", Description: "请输入新密码", HttpCode: BadRequest}
	ErrOriginPassword         = ApiStatus{StatusName: "ORIGIN_PASSWORD_ERROR", Description: "原始密码不正确", HttpCode: BadRequest}
	ErrQQInvalid              = ApiStatus{StatusName: "QQ_INVALID", Description: "qq号不正确", HttpCode: BadRequest}
	SuccessEditCurrentProfile = ApiStatus{StatusName: "SUCCESS_EDIT_CURRENT_PROFILE", Description: "编辑用户信息成功", HttpCode: Ok}
)

func checkQQ(qq int) *ApiStatus {
	if 1e4 <= qq && qq < 1e11 {
		return nil
	}
	return &ErrQQInvalid
}

func (userService *UserService) editUserProfile(req *RequestUserEditCurrentProfile, skipEmailVerify bool) (*ApiStatus, *operation.User) {
	if req.Username == "" && req.Email == "" && req.QQ <= 0 && req.OriginPassword == "" && req.NewPassword == "" && req.AvatarUrl == "" {
		return &ErrIllegalParam, nil
	}
	if req.OriginPassword != "" && req.NewPassword != "" {
		if err := passwordValidator.CheckString(req.NewPassword); err != nil {
			return err, nil
		}
	} else if req.OriginPassword != "" && req.NewPassword == "" {
		return &ErrNewPasswordRequired, nil
	} else if req.OriginPassword == "" && req.NewPassword != "" {
		return &ErrOriginPasswordRequired, nil
	}
	if req.Username != "" {
		if err := usernameValidator.CheckString(req.Username); err != nil {
			return err, nil
		}
	}
	if req.Email != "" {
		if err := emailValidator.CheckString(req.Email); err != nil {
			return err, nil
		}
		if !skipEmailVerify {
			if req.EmailCode <= 0 {
				return &ErrIllegalParam, nil
			}

			if res := userService.verifyEmailCode(req.Email, req.EmailCode, req.Cid); res != nil {
				return res, nil
			}
		}
	}
	if req.QQ > 0 {
		if err := checkQQ(req.QQ); err != nil {
			return err, nil
		}
	}

	user, err := userService.userOperation.GetUserByUid(req.ID)
	if errors.Is(err, operation.ErrUserNotFound) {
		return &ErrUserNotFound, nil
	} else if err != nil {
		return &ErrDatabaseFail, nil
	}

	updateInfo := make(map[string]interface{})

	if req.Username != "" || req.Email != "" {
		exist, _ := userService.userOperation.IsUserIdentifierTaken(nil, 0, req.Username, req.Email)
		if exist {
			return &ErrIdentifierTaken, nil
		}
		if req.Username != "" && req.Username != user.Username {
			user.Username = req.Username
			updateInfo["username"] = req.Username
		}
		if req.Email != "" && req.Email != user.Email {
			user.Email = req.Email
			updateInfo["email"] = req.Email
		}
	}

	if req.QQ > 0 && req.QQ != user.QQ {
		user.QQ = req.QQ
		updateInfo["qq"] = req.QQ
		if req.AvatarUrl == "" && user.AvatarUrl == "" {
			user.AvatarUrl = fmt.Sprintf("https://q2.qlogo.cn/headimg_dl?dst_uin=%d&spec=100", user.QQ)
			updateInfo["avatar_url"] = user.AvatarUrl
		}
	}

	if req.AvatarUrl != "" {
		if user.AvatarUrl != "" && !strings.HasPrefix(user.AvatarUrl, "https://q2.qlogo.cn/") {
			_, err = userService.storeService.DeleteImageFile(user.AvatarUrl)
			if err != nil {
				c.ErrorF("err while delete user old avatar, %v", err)
			}
		}
		user.AvatarUrl = req.AvatarUrl
		updateInfo["avatar_url"] = user.AvatarUrl
	}

	if req.OriginPassword != "" {
		password, err := userService.userOperation.UpdateUserPassword(user, req.OriginPassword, req.NewPassword)
		if errors.Is(err, operation.ErrUserNotFound) {
			return &ErrUserNotFound, nil
		} else if errors.Is(err, operation.ErrOldPassword) {
			return &ErrOriginPassword, nil
		} else if err != nil {
			return &ErrDatabaseFail, nil
		}
		updateInfo["password"] = password
	}

	if err := userService.userOperation.UpdateUserInfo(user, updateInfo); err != nil {
		if errors.Is(err, operation.ErrUserNotFound) {
			return &ErrUserNotFound, nil
		} else {
			return &ErrDatabaseFail, nil
		}
	}

	return nil, user
}

func (userService *UserService) EditCurrentProfile(req *RequestUserEditCurrentProfile) *ApiResponse[ResponseUserEditCurrentProfile] {
	if err, user := userService.editUserProfile(req, false); err != nil {
		return NewApiResponse[ResponseUserEditCurrentProfile](err, Unsatisfied, nil)
	} else {
		return NewApiResponse(&SuccessEditCurrentProfile, Unsatisfied, (*ResponseUserEditCurrentProfile)(user))
	}
}

var (
	SuccessGetProfile = ApiStatus{StatusName: "GET_PROFILE_SUCCESS", Description: "获取用户信息成功", HttpCode: Ok}
)

func (userService *UserService) GetUserProfile(req *RequestUserProfile) *ApiResponse[ResponseUserProfile] {
	if req.TargetUid <= 0 {
		return NewApiResponse[ResponseUserProfile](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseUserProfile](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.UserGetProfile) {
		return NewApiResponse[ResponseUserProfile](&ErrNoPermission, Unsatisfied, nil)
	}
	user, res := CallDBFuncAndCheckError[operation.User, ResponseUserProfile](func() (*operation.User, error) {
		return userService.userOperation.GetUserByUid(req.TargetUid)
	})
	if res != nil {
		return res
	}
	return NewApiResponse(&SuccessGetProfile, Unsatisfied, (*ResponseUserProfile)(user))
}

var (
	SuccessEditUserProfile = ApiStatus{StatusName: "EDIT_USER_PROFILE", Description: "修改用户信息成功", HttpCode: Ok}
)

func (userService *UserService) EditUserProfile(req *RequestUserEditProfile) *ApiResponse[ResponseUserEditProfile] {
	if req.Permission <= 0 {
		return NewApiResponse[ResponseUserEditProfile](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.UserEditBaseInfo) {
		return NewApiResponse[ResponseUserEditProfile](&ErrNoPermission, Unsatisfied, nil)
	}
	req.ID = req.TargetUid
	err, user := userService.editUserProfile(&req.RequestUserEditCurrentProfile, true)
	if err != nil {
		return NewApiResponse[ResponseUserEditProfile](err, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessEditUserProfile, Unsatisfied, (*ResponseUserEditProfile)(user))
}

var SuccessGetUsers = ApiStatus{StatusName: "GET_USER_PAGE", Description: "获取用户信息分页成功", HttpCode: Ok}

func (userService *UserService) GetUserList(req *RequestUserList) *ApiResponse[ResponseUserList] {
	if req.Page <= 0 || req.PageSize <= 0 {
		return NewApiResponse[ResponseUserList](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseUserList](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.UserShowList) {
		return NewApiResponse[ResponseUserList](&ErrNoPermission, Unsatisfied, nil)
	}
	users, total, err := userService.userOperation.GetUsers(req.Page, req.PageSize)
	if err != nil {
		return NewApiResponse[ResponseUserList](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessGetUsers, Unsatisfied, &ResponseUserList{
		Items:    users,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	})
}

var SuccessGetControllers = ApiStatus{StatusName: "GET_CONTROLLER_PAGE", Description: "获取管制员信息分页成功", HttpCode: Ok}

func (userService *UserService) GetControllerList(req *RequestControllerList) *ApiResponse[ResponseControllerList] {
	if req.Page <= 0 || req.PageSize <= 0 {
		return NewApiResponse[ResponseControllerList](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseControllerList](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.UserShowList) {
		return NewApiResponse[ResponseControllerList](&ErrNoPermission, Unsatisfied, nil)
	}
	users, total, err := userService.userOperation.GetControllers(req.Page, req.PageSize)
	if err != nil {
		return NewApiResponse[ResponseControllerList](&ErrDatabaseFail, Unsatisfied, nil)
	}
	return NewApiResponse(&SuccessGetControllers, Unsatisfied, &ResponseControllerList{
		Items:    users,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	})
}

var (
	ErrPermissionNodeNotExists = ApiStatus{StatusName: "PERMISSION_NODE_NOT_EXISTS", Description: "无效权限节点", HttpCode: BadRequest}
	SuccessEditUserPermission  = ApiStatus{StatusName: "EDIT_USER_PERMISSION", Description: "编辑用户权限成功", HttpCode: Ok}
)

func (userService *UserService) EditUserPermission(req *RequestUserEditPermission) *ApiResponse[ResponseUserEditPermission] {
	if req.Uid <= 0 {
		return NewApiResponse[ResponseUserEditPermission](&ErrIllegalParam, Unsatisfied, nil)
	}
	user, targetUser, res := GetUsersAndCheckPermission[ResponseUserEditPermission](userService.userOperation, req.Uid, req.TargetUid, operation.UserEditPermission)
	if res != nil {
		return res
	}
	permission := operation.Permission(user.Permission)
	targetPermission := operation.Permission(targetUser.Permission)
	for key, value := range req.Permissions {
		if per, ok := operation.PermissionMap[key]; ok {
			if !permission.HasPermission(per) {
				return NewApiResponse[ResponseUserEditPermission](&ErrNoPermission, Unsatisfied, nil)
			}
			if value, ok := value.(bool); ok {
				if value {
					targetPermission.Grant(per)
				} else {
					targetPermission.Revoke(per)
				}
			} else {
				return NewApiResponse[ResponseUserEditPermission](&ErrIllegalParam, Unsatisfied, nil)
			}
		} else {
			return NewApiResponse[ResponseUserEditPermission](&ErrPermissionNodeNotExists, Unsatisfied, nil)
		}
	}

	if _, res := CallDBFuncAndCheckError[interface{}, ResponseUserEditPermission](func() (*interface{}, error) {
		return nil, userService.userOperation.UpdateUserPermission(targetUser, targetPermission)
	}); res != nil {
		return res
	}

	if userService.config.Email.Template.EnablePermissionChangeEmail {
		if err := userService.emailService.SendPermissionChangeEmail(targetUser, user); err != nil {
			c.ErrorF("SendPermissionChangeEmail Failed: %v", err)
		}
	}

	return NewApiResponse(&SuccessEditUserPermission, Unsatisfied, (*ResponseUserEditPermission)(user))
}

var (
	ErrSameRating         = ApiStatus{StatusName: "SAME_RATING", Description: "用户已是该权限", HttpCode: BadRequest}
	SuccessEditUserRating = ApiStatus{StatusName: "EDIT_USER_RATING", Description: "编辑用户管制权限成功", HttpCode: Ok}
)

func (userService *UserService) EditUserRating(req *RequestUserEditRating) *ApiResponse[ResponseUserEditRating] {
	if req.Uid <= 0 || req.Rating < fsd.Ban.Index() || req.Rating > fsd.Administrator.Index() {
		return NewApiResponse[ResponseUserEditRating](&ErrIllegalParam, Unsatisfied, nil)
	}
	user, targetUser, res := GetUsersAndCheckPermission[ResponseUserEditRating](userService.userOperation, req.Uid, req.TargetUid, operation.UserEditRating)
	if res != nil {
		return res
	}
	oldRating := fsd.Rating(targetUser.Rating)
	newRating := fsd.Rating(req.Rating)
	if oldRating == newRating {
		return NewApiResponse[ResponseUserEditRating](&ErrSameRating, Unsatisfied, nil)
	}

	if _, res := CallDBFuncAndCheckError[interface{}, ResponseUserEditRating](func() (*interface{}, error) {
		return nil, userService.userOperation.UpdateUserRating(targetUser, newRating.Index())
	}); res != nil {
		return res
	}

	if userService.config.Email.Template.EnableRatingChangeEmail {
		if err := userService.emailService.SendRatingChangeEmail(targetUser, user, oldRating, newRating); err != nil {
			c.ErrorF("SendRatingChangeEmail Failed: %v", err)
		}
	}

	return NewApiResponse(&SuccessEditUserRating, Unsatisfied, (*ResponseUserEditRating)(user))
}

var SuccessGetUserHistory = ApiStatus{StatusName: "GET_USER_HISTORY", Description: "成功获取用户历史数据", HttpCode: Ok}

func (userService *UserService) GetUserHistory(req *RequestGetUserHistory) *ApiResponse[ResponseGetUserHistory] {
	if req.Cid <= 0 {
		return NewApiResponse[ResponseGetUserHistory](&ErrIllegalParam, Unsatisfied, nil)
	}

	user, res := CallDBFuncAndCheckError[operation.User, ResponseGetUserHistory](func() (*operation.User, error) {
		return userService.userOperation.GetUserByCid(req.Cid)
	})
	if res != nil {
		return res
	}

	userHistory, res := CallDBFuncAndCheckError[operation.UserHistory, ResponseGetUserHistory](func() (*operation.UserHistory, error) {
		return userService.historyOperation.GetUserHistory(req.Cid)
	})
	if res != nil {
		return res
	}

	return NewApiResponse(&SuccessGetUserHistory, Unsatisfied, &ResponseGetUserHistory{
		TotalPilotTime: user.TotalPilotTime,
		TotalAtcTime:   user.TotalAtcTime,
		UserHistory:    userHistory,
	})
}

var SuccessGetToken = ApiStatus{StatusName: "GET_TOKEN", Description: "成功刷新秘钥", HttpCode: Ok}

func (userService *UserService) GetTokenWithFlushToken(req *RequestGetToken) *ApiResponse[ResponseGetToken] {
	if !req.FlushToken {
		return NewApiResponse[ResponseGetToken](&ErrIllegalParam, Unsatisfied, nil)
	}

	user, res := CallDBFuncAndCheckError[operation.User, ResponseGetToken](func() (*operation.User, error) {
		return userService.userOperation.GetUserByUid(req.Uid)
	})

	if res != nil {
		return res
	}

	var flushToken string
	if req.ExpiresAt.Add(-2 * userService.config.JWT.ExpiresDuration).After(time.Now()) {
		flushToken = ""
	} else {
		flushToken = NewClaims(userService.config.JWT, user, true).GenerateKey()
	}

	token := NewClaims(userService.config.JWT, user, false)
	return NewApiResponse(&SuccessGetToken, Unsatisfied, &ResponseGetToken{
		User:       user,
		Token:      token.GenerateKey(),
		FlushToken: flushToken,
	})
}
