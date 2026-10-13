// Package mailer 邮件/验证码发送抽象，供第三方复用：
//   - Webhook：自有发码平台（HTTP POST JSON，最小契约见 server/README.md）
//   - SMTP：标准邮箱（465 隐式 TLS / 587 STARTTLS 自适应）
//   - Console：开发默认，验证码打印到日志
//
// 优先级：webhook > SMTP > Console（New 按配置自动选择）。
package mailer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

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

// Webhook 把验证码投递到自有发码平台（HTTP POST JSON）。
// 请求体 {"email":"...","code":"...","purpose":"register|reset"}；
// 可选 Authorization: Bearer <Token>；2xx 视为发送成功。
type Webhook struct {
	URL    string
	Token  string
	Log    *log.Logger
	Client *http.Client
}

// SendCode POST 到发码平台。
func (w *Webhook) SendCode(to, purpose, code string) error {
	payload, err := json.Marshal(map[string]string{
		"email":   to,
		"code":    code,
		"purpose": purpose,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, w.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if w.Token != "" {
		req.Header.Set("Authorization", "Bearer "+w.Token)
	}
	cli := w.Client
	if cli == nil {
		cli = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("mailer webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("mailer webhook: HTTP %d", resp.StatusCode)
	}
	return nil
}

// SMTP 通过标准 SMTP 发信。Port 465 走隐式 TLS，其余端口先明文连接、
// 服务器支持则 STARTTLS；User 为空时不做认证（内网中继）。
type SMTP struct {
	Host string
	Port string // 465 | 587 | 25 ...
	User string
	Pass string
	From string // 发件人地址
	Log  *log.Logger
}

// SendCode 发送验证码邮件。
func (s *SMTP) SendCode(to, purpose, code string) error {
	subject := "轻音 LiteTalk 验证码"
	if purpose == "reset" {
		subject = "轻音 LiteTalk 密码重置验证码"
	}
	msg := strings.Join([]string{
		"From: " + s.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		fmt.Sprintf("验证码：%s（10 分钟内有效，请勿转发）", code),
	}, "\r\n")
	return s.send(to, []byte(msg))
}

func (s *SMTP) send(to string, msg []byte) error {
	addr := net.JoinHostPort(s.Host, s.Port)
	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	if s.User == "" {
		auth = nil
	}
	if s.Port == "465" {
		// 隐式 TLS（SMTPS）。
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.Host})
		if err != nil {
			return fmt.Errorf("mailer smtp: %w", err)
		}
		defer conn.Close()
		cl, err := smtp.NewClient(conn, s.Host)
		if err != nil {
			return err
		}
		defer cl.Quit()
		if auth != nil {
			if err := cl.Auth(auth); err != nil {
				return err
			}
		}
		return deliver(cl, s.From, to, msg)
	}
	// 明文连接 + STARTTLS（服务器支持时）。
	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}

func deliver(cl *smtp.Client, from, to string, msg []byte) error {
	if err := cl.Mail(from); err != nil {
		return err
	}
	if err := cl.Rcpt(to); err != nil {
		return err
	}
	w, err := cl.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

// Config 发信配置（一般全部来自环境变量）。
type Config struct {
	WebhookURL   string // LITETALK_MAILER_WEBHOOK
	WebhookToken string // LITETALK_MAILER_WEBHOOK_TOKEN
	SMTPHost     string // LITETALK_SMTP_HOST
	SMTPPort     string // LITETALK_SMTP_PORT
	SMTPUser     string // LITETALK_SMTP_USER
	SMTPPass     string // LITETALK_SMTP_PASS
	SMTPFrom     string // LITETALK_SMTP_FROM
}

// New 按优先级 webhook > SMTP > Console 构造发送器，返回实现与名称。
// SMTP 配置不全时降级 Console 并在日志提示。
func New(cfg Config, lg *log.Logger) (Mailer, string) {
	if lg == nil {
		lg = log.Default()
	}
	if cfg.WebhookURL != "" {
		return &Webhook{URL: cfg.WebhookURL, Token: cfg.WebhookToken, Log: lg}, "webhook"
	}
	if cfg.SMTPHost != "" && cfg.SMTPFrom != "" {
		port := cfg.SMTPPort
		if port == "" {
			port = "587"
		}
		return &SMTP{
			Host: cfg.SMTPHost, Port: port,
			User: cfg.SMTPUser, Pass: cfg.SMTPPass, From: cfg.SMTPFrom,
			Log: lg,
		}, "smtp"
	}
	if cfg.SMTPHost != "" || cfg.SMTPFrom != "" || cfg.SMTPUser != "" {
		lg.Println("[mailer] SMTP 配置不全（需要 HOST 与 FROM），已降级为控制台打印")
	}
	return &Console{Log: lg}, "console"
}
