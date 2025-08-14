// Package interfaces
package interfaces

type UserServiceInterface interface {
	RegisterUser(req *RequestRegisterUser) *ApiResponse[ResponseRegisterUser]
	UserLogin(req *RequestUserLogin) *ApiResponse[ResponseUserLogin]
	CheckAvailability(req *RequestUserAvailability) *ApiResponse[ResponseUserAvailability]
	GetCurrentProfile(req *RequestUserCurrentProfile) *ApiResponse[ResponseUserCurrentProfile]
	EditCurrentProfile(req *RequestUserEditCurrentProfile) *ApiResponse[ResponseUserEditCurrentProfile]
	GetUserProfile(req *RequestUserProfile) *ApiResponse[ResponseUserProfile]
	EditUserProfile(req *RequestUserEditProfile) *ApiResponse[ResponseUserEditProfile]
	GetUserList(req *RequestUserList) *ApiResponse[ResponseUserList]
}

type RequestRegisterUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Cid       int    `json:"cid"`
	EmailCode int    `json:"email_code"`
}

type ResponseRegisterUser struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

type RequestUserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseUserLogin struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

type RequestUserAvailability struct {
	Username string `query:"username"`
	Email    string `query:"email"`
	Cid      string `query:"cid"`
}

type ResponseUserAvailability bool

type RequestUserCurrentProfile struct {
	Uid uint
}

type ResponseUserCurrentProfile struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Cid            int    `json:"cid"`
	QQ             int    `json:"qq"`
	Rating         int    `json:"rating"`
	TotalPilotTime int    `json:"total_pilot_time"`
	TotalAtcTime   int    `json:"total_atc_time"`
	Permission     int64  `json:"permission"`
}

type RequestUserEditCurrentProfile struct {
	ID             uint   `json:"id"`
	Cid            int    `json:"cid"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	EmailCode      int    `json:"email_code"`
	QQ             int    `json:"qq"`
	OriginPassword string `json:"origin_password"`
	NewPassword    string `json:"new_password"`
}

type ResponseUserEditCurrentProfile struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Cid            int    `json:"cid"`
	QQ             int    `json:"qq"`
	Rating         int    `json:"rating"`
	TotalPilotTime int    `json:"total_pilot_time"`
	TotalAtcTime   int    `json:"total_atc_time"`
	Permission     int64  `json:"permission"`
}

type RequestUserProfile struct {
	JwtHeader
	TargetUid uint `param:"uid"`
}

type ResponseUserProfile ResponseUserCurrentProfile

type RequestUserList struct {
	JwtHeader
	Page     int `query:"page_number"`
	PageSize int `query:"page_size"`
}

type ResponseUserList struct {
	Items    []ResponseUserCurrentProfile `json:"items"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
	Total    int64                        `json:"total"`
}

type RequestUserEditProfile struct {
	JwtHeader
	TargetUid uint `param:"uid"`
	RequestUserEditCurrentProfile
}

type ResponseUserEditProfile ResponseUserCurrentProfile
