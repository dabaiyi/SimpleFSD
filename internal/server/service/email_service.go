// Package service
package service

import (
	"errors"
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"gopkg.in/gomail.v2"
	"html/template"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type EmailCode struct {
	code     int
	cid      int
	sendTime time.Time
}

type EmailManager struct {
	emailCodes   map[string]EmailCode
	lastSendTime map[string]time.Time
}

type EmailVerifyTemplateData struct {
	Cid     string
	Code    string
	Expired string
}

var (
	emailManager *EmailManager
	config       *c.Config
	once         sync.Once
)

func GetEmailManager() *EmailManager {
	once.Do(func() {
		config, _ = c.GetConfig()
		emailManager = &EmailManager{
			emailCodes:   make(map[string]EmailCode),
			lastSendTime: make(map[string]time.Time),
		}
	})
	return emailManager
}

func renderTemplateSafe(template *template.Template, data interface{}) (string, error) {
	var sb strings.Builder
	if err := template.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
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

func (evt *EmailVerifyTemplateData) render() (string, error) {
	if config.Server.HttpServer.Email.Template.EmailVerifyTemplate == nil {
		return "", ErrTemplateNotInitialized
	}

	return renderTemplateSafe(config.Server.HttpServer.Email.Template.EmailVerifyTemplate, evt)
}

func (em *EmailManager) VerifyCode(email string, code int, cid int) error {
	emailCode, ok := em.emailCodes[email]
	if !ok {
		return ErrEmailCodeNotFound
	}

	if time.Since(emailCode.sendTime) > config.Server.HttpServer.Email.VerifyExpiredDuration {
		return ErrEmailCodeExpired
	}

	if emailCode.code != code {
		return ErrInvalidEmailCode
	}

	if emailCode.cid != cid {
		return ErrCidMismatch
	}

	delete(em.emailCodes, email)
	return nil
}
func (em *EmailManager) SendEmailVerifyCode(email string, cid int) error {
	if lastSendTime, ok := em.lastSendTime[email]; ok {
		if time.Since(lastSendTime) < config.Server.HttpServer.Email.SendDuration {
			return ErrEmailSendInterval
		}
	}
	code := rand.Intn(9e5) + 1e5
	emailCode := EmailCode{code: code, cid: cid, sendTime: time.Now()}
	data := &EmailVerifyTemplateData{
		Cid:     strconv.Itoa(cid),
		Code:    strconv.Itoa(code),
		Expired: strconv.Itoa(int(config.Server.HttpServer.Email.VerifyExpiredDuration.Minutes())),
	}

	message, err := data.render()
	if err != nil {
		c.WarnF("Error rendering email verification template: %v", err)
		return ErrRenderingTemplate
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.Server.HttpServer.Email.Username)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "您的验证码")
	m.SetBody("text/html", message)

	em.emailCodes[email] = emailCode
	em.lastSendTime[email] = time.Now()

	c.InfoF("Sending email verification code(%d) to %s(%d)", code, email, cid)

	if err := config.Server.HttpServer.Email.EmailServer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

type EmailVerifyCodeData struct {
	Email string `json:"email"`
	Cid   int    `json:"cid"`
}

type EmailVerifyCodeResponse struct {
	Email string `json:"email"`
}

var (
	SendEmailSuccess  = ApiStatus{"SEND_EMAIL_SUCCESS", "邮件发送成功", Ok}
	ErrRenderTemplate = ApiStatus{"RENDER_TEMPLATE_ERROR", "发送失败", ServerInternalError}
	ErrSendEmail      = ApiStatus{"EMAIL_SEND_ERROR", "发送失败", ServerInternalError}
)

func (evc *EmailVerifyCodeData) SendEmailVerifyCode() *ApiResponse[EmailVerifyCodeResponse] {
	if evc.Email == "" || evc.Cid <= 0 {
		return NewApiResponse[EmailVerifyCodeResponse](&ErrIllegalParam, Unsatisfied, nil)
	}
	err := emailManager.SendEmailVerifyCode(evc.Email, evc.Cid)
	if err == nil {
		return NewApiResponse(&SendEmailSuccess, Unsatisfied, &EmailVerifyCodeResponse{evc.Email})
	}
	if errors.Is(err, ErrEmailSendInterval) {
		return NewApiResponse[EmailVerifyCodeResponse](&ApiStatus{
			"EMAIL_SEND_INTERVAL",
			fmt.Sprintf("邮件已发送, 请在%d秒后重试",
				int(config.Server.HttpServer.Email.SendDuration.Seconds())),
			BadRequest,
		}, Unsatisfied, nil)
	}
	if errors.Is(err, ErrRenderingTemplate) {
		return NewApiResponse[EmailVerifyCodeResponse](&ErrRenderTemplate, Unsatisfied, nil)
	}
	return NewApiResponse[EmailVerifyCodeResponse](&ErrSendEmail, Unsatisfied, nil)
}
