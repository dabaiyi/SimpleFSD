// Package service
package service

import (
	"errors"
	"fmt"
	"github.com/half-nothing/simple-fsd/internal/interfaces/config"
	"github.com/half-nothing/simple-fsd/internal/interfaces/fsd"
	"github.com/half-nothing/simple-fsd/internal/interfaces/log"
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"gopkg.in/gomail.v2"
	"html/template"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	emailService *EmailService
	once         sync.Once
)

type EmailService struct {
	logger       log.LoggerInterface
	emailCodes   map[string]EmailCode
	lastSendTime map[string]time.Time
	config       *config.EmailConfig
}

type EmailCode struct {
	code     int
	cid      int
	sendTime time.Time
}

type EmailVerifyTemplateData struct {
	Cid     string
	Code    string
	Expired string
}

type EmailPermissionChangeData struct {
	Cid      string
	Operator string
	Contact  string
}

type EmailRatingChangeData struct {
	Cid      string
	NewValue string
	OldValue string
	Operator string
	Contact  string
}

type EmailKickedFromServerData struct {
	Cid      string
	Time     string
	Reason   string
	Operator string
	Contact  string
}

func NewEmailService(logger log.LoggerInterface, config *config.EmailConfig) *EmailService {
	once.Do(func() {
		emailService = &EmailService{
			logger:       logger,
			config:       config,
			emailCodes:   make(map[string]EmailCode),
			lastSendTime: make(map[string]time.Time),
		}
	})
	return emailService
}

var (
	ErrEmailSendInterval      = errors.New("email send interval")
	ErrRenderingTemplate      = errors.New("error rendering template")
	ErrTemplateNotInitialized = errors.New("error template not initialized")
	ErrEmailCodeNotFound      = errors.New("email code not found")
	ErrEmailCodeExpired       = errors.New("email code expired")
	ErrInvalidEmailCode       = errors.New("invalid email code")
	ErrCidMismatch            = errors.New("cid mismatch")
)

func (emailService *EmailService) RenderTemplate(template *template.Template, data interface{}) (string, error) {
	if template == nil {
		return "", ErrTemplateNotInitialized
	}
	var sb strings.Builder
	if err := template.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (emailService *EmailService) VerifyCode(email string, code int, cid int) error {
	if emailService.config.EmailServer == nil {
		return nil
	}
	email = strings.ToLower(email)
	emailCode, ok := emailService.emailCodes[email]
	if !ok {
		return ErrEmailCodeNotFound
	}

	if time.Since(emailCode.sendTime) > emailService.config.VerifyExpiredDuration {
		return ErrEmailCodeExpired
	}

	if emailCode.code != code {
		return ErrInvalidEmailCode
	}

	if emailCode.cid != cid {
		return ErrCidMismatch
	}

	delete(emailService.emailCodes, email)
	return nil
}

func (emailService *EmailService) SendEmailCode(email string, cid int) error {
	if emailService.config.EmailServer == nil {
		return nil
	}
	email = strings.ToLower(email)
	if lastSendTime, ok := emailService.lastSendTime[email]; ok {
		if time.Since(lastSendTime) < emailService.config.SendDuration {
			return ErrEmailSendInterval
		}
	}
	code := rand.Intn(9e5) + 1e5
	emailCode := EmailCode{code: code, cid: cid, sendTime: time.Now()}
	data := &EmailVerifyTemplateData{
		Cid:     fmt.Sprintf("%04d", cid),
		Code:    strconv.Itoa(code),
		Expired: strconv.Itoa(int(emailService.config.VerifyExpiredDuration.Minutes())),
	}

	message, err := emailService.RenderTemplate(emailService.config.Template.EmailVerifyTemplate, data)
	if err != nil {
		emailService.logger.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "您的验证码")
	m.SetBody("text/html", message)

	emailService.emailCodes[email] = emailCode
	emailService.lastSendTime[email] = time.Now()

	emailService.logger.InfoF("Sending email verification code(%d) to %s(%d)", code, email, cid)

	return emailService.config.EmailServer.DialAndSend(m)
}

func (emailService *EmailService) SendPermissionChangeEmail(user *operation.User, operator *operation.User) error {
	if emailService.config.EmailServer == nil {
		return nil
	}
	email := strings.ToLower(user.Email)
	data := &EmailPermissionChangeData{
		Cid:      fmt.Sprintf("%04d", user.Cid),
		Operator: fmt.Sprintf("%04d", operator.Cid),
		Contact:  operator.Email,
	}
	message, err := emailService.RenderTemplate(emailService.config.Template.PermissionChangeTemplate, data)
	if err != nil {
		emailService.logger.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "管理权限变更通知")
	m.SetBody("text/html", message)

	emailService.logger.InfoF("Sending permission change email to %s(%d)", email, user.Cid)

	return emailService.config.EmailServer.DialAndSend(m)
}

func (emailService *EmailService) SendRatingChangeEmail(user *operation.User, operator *operation.User, oldRating, newRating fsd.Rating) error {
	if emailService.config.EmailServer == nil {
		return nil
	}
	email := strings.ToLower(user.Email)
	data := &EmailRatingChangeData{
		Cid:      strconv.Itoa(user.Cid),
		OldValue: oldRating.String(),
		NewValue: newRating.String(),
		Operator: fmt.Sprintf("%04d", operator.Cid),
		Contact:  operator.Email,
	}
	message, err := emailService.RenderTemplate(emailService.config.Template.ATCRatingChangeTemplate, data)
	if err != nil {
		emailService.logger.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "管制权限变更通知")
	m.SetBody("text/html", message)

	emailService.logger.InfoF("Sending rating change email to %s(%d)", email, user.Cid)

	return emailService.config.EmailServer.DialAndSend(m)
}

func (emailService *EmailService) SendKickedFromServerEmail(user *operation.User, operator *operation.User, reason string) error {
	if emailService.config.EmailServer == nil {
		return nil
	}
	email := strings.ToLower(user.Email)
	data := &EmailKickedFromServerData{
		Cid:      strconv.Itoa(user.Cid),
		Time:     time.Now().Format(time.DateTime),
		Reason:   reason,
		Operator: fmt.Sprintf("%04d", operator.Cid),
		Contact:  operator.Email,
	}
	message, err := emailService.RenderTemplate(emailService.config.Template.KickedFromServerTemplate, data)
	if err != nil {
		emailService.logger.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "踢出服务器通知")
	m.SetBody("text/html", message)

	emailService.logger.InfoF("Sending kick message email to %s(%d)", email, user.Cid)

	return emailService.config.EmailServer.DialAndSend(m)
}

var (
	SendEmailSuccess  = ApiStatus{StatusName: "SEND_EMAIL_SUCCESS", Description: "邮件发送成功", HttpCode: Ok}
	ErrRenderTemplate = ApiStatus{StatusName: "RENDER_TEMPLATE_ERROR", Description: "发送失败", HttpCode: ServerInternalError}
	ErrSendEmail      = ApiStatus{StatusName: "EMAIL_SEND_ERROR", Description: "发送失败", HttpCode: ServerInternalError}
)

func (emailService *EmailService) SendEmailVerifyCode(req *RequestEmailVerifyCode) *ApiResponse[ResponseEmailVerifyCode] {
	if emailService.config.EmailServer == nil {
		return NewApiResponse(&SendEmailSuccess, Unsatisfied, &ResponseEmailVerifyCode{Email: req.Email})
	}
	if req.Email == "" || req.Cid <= 0 {
		return NewApiResponse[ResponseEmailVerifyCode](&ErrIllegalParam, Unsatisfied, nil)
	}
	err := emailService.SendEmailCode(req.Email, req.Cid)
	if err == nil {
		return NewApiResponse(&SendEmailSuccess, Unsatisfied, &ResponseEmailVerifyCode{Email: req.Email})
	}
	if errors.Is(err, ErrEmailSendInterval) {
		return NewApiResponse[ResponseEmailVerifyCode](&ApiStatus{
			StatusName: "EMAIL_SEND_INTERVAL",
			Description: fmt.Sprintf("邮件已发送, 请在%d秒后重试",
				int(emailService.config.SendDuration.Seconds())),
			HttpCode: BadRequest,
		}, Unsatisfied, nil)
	}
	if errors.Is(err, ErrRenderingTemplate) {
		return NewApiResponse[ResponseEmailVerifyCode](&ErrRenderTemplate, Unsatisfied, nil)
	}
	return NewApiResponse[ResponseEmailVerifyCode](&ErrSendEmail, Unsatisfied, nil)
}
