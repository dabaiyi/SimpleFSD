// Package config
package config

import (
	"errors"
	"github.com/half-nothing/simple-fsd/internal/interfaces/global"
	"github.com/half-nothing/simple-fsd/internal/interfaces/log"
	"html/template"
)

type EmailTemplateConfig struct {
	EmailVerifyTemplateFile      string             `json:"email_verify_template_file"`
	EmailVerifyTemplate          *template.Template `json:"-"`
	ATCRatingChangeTemplateFile  string             `json:"atc_rating_change_template_file"`
	ATCRatingChangeTemplate      *template.Template `json:"-"`
	EnableRatingChangeEmail      bool               `json:"enable_rating_change_email"`
	PermissionChangeTemplateFile string             `json:"permission_change_template_file"`
	PermissionChangeTemplate     *template.Template `json:"-"`
	EnablePermissionChangeEmail  bool               `json:"enable_permission_change_email"`
	KickedFromServerTemplateFile string             `json:"kicked_from_server_template_file"`
	KickedFromServerTemplate     *template.Template `json:"-"`
	EnableKickedFromServerEmail  bool               `json:"enable_kicked_from_server_email"`
}

func defaultEmailTemplateConfig() *EmailTemplateConfig {
	return &EmailTemplateConfig{
		EmailVerifyTemplateFile:      "template/email_verify.template",
		ATCRatingChangeTemplateFile:  "template/atc_rating_change.template",
		EnableRatingChangeEmail:      true,
		PermissionChangeTemplateFile: "template/permission_change.template",
		EnablePermissionChangeEmail:  true,
		KickedFromServerTemplateFile: "template/kicked_from_server.template",
		EnableKickedFromServerEmail:  true,
	}
}

func (config *EmailTemplateConfig) checkValid(logger log.LoggerInterface) *ValidResult {
	if bytes, err := cachedContent(logger, config.EmailVerifyTemplateFile, global.EmailVerifyTemplateFileUrl); err != nil {
		return ValidFailWith(errors.New("fail to load email_verify_template_file"), err)
	} else if parse, err := template.New("email_verify").Parse(string(bytes)); err != nil {
		return ValidFailWith(errors.New("fail to parse email_verify_template"), err)
	} else {
		config.EmailVerifyTemplate = parse
	}

	if config.EnableRatingChangeEmail {
		if bytes, err := cachedContent(logger, config.ATCRatingChangeTemplateFile, global.ATCRatingChangeTemplateFileUrl); err != nil {
			return ValidFailWith(errors.New("fail to load atc_rating_change_template_file"), err)
		} else if parse, err := template.New("atc_rating_change").Parse(string(bytes)); err != nil {
			return ValidFailWith(errors.New("fail to parse atc_rating_change_template"), err)
		} else {
			config.ATCRatingChangeTemplate = parse
		}
	}

	if config.EnablePermissionChangeEmail {
		if bytes, err := cachedContent(logger, config.PermissionChangeTemplateFile, global.PermissionChangeTemplateFileUrl); err != nil {
			return ValidFailWith(errors.New("fail to load permission_change_template_file"), err)
		} else if parse, err := template.New("permission_change").Parse(string(bytes)); err != nil {
			return ValidFailWith(errors.New("fail to parse permission_change_template"), err)
		} else {
			config.PermissionChangeTemplate = parse
		}
	}

	if config.EnableKickedFromServerEmail {
		if bytes, err := cachedContent(logger, config.KickedFromServerTemplateFile, global.KickedFromServerTemplateFileUrl); err != nil {
			return ValidFailWith(errors.New("fail to load permission_change_template_file"), err)
		} else if parse, err := template.New("kicked_from_server").Parse(string(bytes)); err != nil {
			return ValidFailWith(errors.New("fail to parse permission_change_template"), err)
		} else {
			config.KickedFromServerTemplate = parse
		}
	}

	return ValidPass()
}
