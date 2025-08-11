package main

import (
	"gopkg.in/gomail.v2"
)

func main() {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@half-nothing.cn")
	m.SetHeader("To", "half_nothing@163.com")
	m.SetHeader("Subject", "您的验证码")
	m.SetBody("text/html", "<p>您的邮箱验证码是456324, 请勿泄露给他人</p>\n<p>验证码3分钟内有效, 请尽快使用</p>") // 支持HTML

	d := gomail.NewDialer("gz-smtp.qcloudmail.com", 465, "noreply@half-nothing.cn", "3suOBsIV8yiz") // QQ邮箱示例

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
