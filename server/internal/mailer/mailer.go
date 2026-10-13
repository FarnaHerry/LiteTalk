// Package mailer 邮件发送抽象。M3 提供控制台实现（dev 打印验证码）；
// 阿里云邮件推送接入留 TODO（需 LITETALK_SMTP_* 凭据）。
package mailer

import "log"

// Mailer 邮件发送接口。
type Mailer interface {
	// SendCode 发送验证码邮件。
	SendCode(to, purpose, code string) error
}

// Console 是开发用实现：验证码打印到日志，不真发邮件。
type Console struct {
	Log *log.Logger
}

// SendCode 打印验证码。
func (c *Console) SendCode(to, purpose, code string) error {
	lg := c.Log
	if lg == nil {
		lg = log.Default()
	}
	lg.Printf("[mailer/console] to=%s purpose=%s code=%s", to, purpose, code)
	return nil
}

// TODO(M3 上线): 阿里云邮件推送 / SMTP 实现（SPF/DKIM 配置见部署文档）。
