// Package service
package service

import (
	"errors"
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	. "github.com/half-nothing/fsd-server/internal/server/service/interfaces"
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
	emailCodes   map[string]EmailCode
	lastSendTime map[string]time.Time
	config       *c.Config
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

func NewEmailService(config *c.Config) *EmailService {
	once.Do(func() {
		emailService = &EmailService{
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
	email = strings.ToLower(email)
	emailCode, ok := emailService.emailCodes[email]
	if !ok {
		return ErrEmailCodeNotFound
	}

	if time.Since(emailCode.sendTime) > emailService.config.Server.HttpServer.Email.VerifyExpiredDuration {
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
	email = strings.ToLower(email)
	if lastSendTime, ok := emailService.lastSendTime[email]; ok {
		if time.Since(lastSendTime) < emailService.config.Server.HttpServer.Email.SendDuration {
			return ErrEmailSendInterval
		}
	}
	code := rand.Intn(9e5) + 1e5
	emailCode := EmailCode{code: code, cid: cid, sendTime: time.Now()}
	data := &EmailVerifyTemplateData{
		Cid:     strconv.Itoa(cid),
		Code:    strconv.Itoa(code),
		Expired: strconv.Itoa(int(emailService.config.Server.HttpServer.Email.VerifyExpiredDuration.Minutes())),
	}

	message, err := emailService.RenderTemplate(emailService.config.Server.HttpServer.Email.Template.EmailVerifyTemplate, data)
	if err != nil {
		c.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Server.HttpServer.Email.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "您的验证码")
	m.SetBody("text/html", message)

	emailService.emailCodes[email] = emailCode
	emailService.lastSendTime[email] = time.Now()

	c.InfoF("Sending email verification code(%d) to %s(%d)", code, email, cid)

	return emailService.config.Server.HttpServer.Email.EmailServer.DialAndSend(m)
}

func (emailService *EmailService) SendPermissionChangeEmail(user *database.User, operator *database.User) error {
	email := strings.ToLower(user.Email)
	data := &EmailPermissionChangeData{
		Cid:      strconv.Itoa(user.Cid),
		Operator: strconv.Itoa(operator.Cid),
		Contact:  operator.Email,
	}
	message, err := emailService.RenderTemplate(emailService.config.Server.HttpServer.Email.Template.PermissionChangeTemplate, data)
	if err != nil {
		c.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Server.HttpServer.Email.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "管理权限变更通知")
	m.SetBody("text/html", message)

	return emailService.config.Server.HttpServer.Email.EmailServer.DialAndSend(m)
}

func (emailService *EmailService) SendRatingChangeEmail(user *database.User, operator *database.User, oldRating, newRating packet.Rating) error {
	email := strings.ToLower(user.Email)
	data := &EmailRatingChangeData{
		Cid:      strconv.Itoa(user.Cid),
		OldValue: oldRating.String(),
		NewValue: newRating.String(),
		Operator: strconv.Itoa(operator.Cid),
		Contact:  operator.Email,
	}
	message, err := emailService.RenderTemplate(emailService.config.Server.HttpServer.Email.Template.ATCRatingChangeTemplate, data)
	if err != nil {
		c.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailService.config.Server.HttpServer.Email.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "管制权限变更通知")
	m.SetBody("text/html", message)

	return emailService.config.Server.HttpServer.Email.EmailServer.DialAndSend(m)
}

var (
	SendEmailSuccess  = ApiStatus{StatusName: "SEND_EMAIL_SUCCESS", Description: "邮件发送成功", HttpCode: Ok}
	ErrRenderTemplate = ApiStatus{StatusName: "RENDER_TEMPLATE_ERROR", Description: "发送失败", HttpCode: ServerInternalError}
	ErrSendEmail      = ApiStatus{StatusName: "EMAIL_SEND_ERROR", Description: "发送失败", HttpCode: ServerInternalError}
)

func (emailService *EmailService) SendEmailVerifyCode(req *RequestEmailVerifyCode) *ApiResponse[ResponseEmailVerifyCode] {
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
				int(emailService.config.Server.HttpServer.Email.SendDuration.Seconds())),
			HttpCode: BadRequest,
		}, Unsatisfied, nil)
	}
	if errors.Is(err, ErrRenderingTemplate) {
		return NewApiResponse[ResponseEmailVerifyCode](&ErrRenderTemplate, Unsatisfied, nil)
	}
	return NewApiResponse[ResponseEmailVerifyCode](&ErrSendEmail, Unsatisfied, nil)
}
