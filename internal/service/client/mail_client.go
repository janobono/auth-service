package client

import (
	"bytes"
	"github.com/janobono/auth-service/internal/config"
	"gopkg.in/gomail.v2"
	"html/template"
	"log/slog"
	"os"
	"sync"
	"time"
)

type MailData struct {
	From        string
	ReplyTo     string
	Recipients  []string
	Cc          []string
	Subject     string
	Content     *MailContentData
	Attachments map[string]string // filename -> file path
}

type MailContentData struct {
	Title string
	Lines []string
	Link  *MailLinkData
}

type MailLinkData struct {
	Href string
	Text string
}

type MailClient interface {
	SendEmail(data *MailData)
}

type mailClient struct {
	mailConfig   *config.MailConfig
	mailTemplate *template.Template
	mu           sync.RWMutex
}

var _ MailClient = (*mailClient)(nil)

func NewMailClient(mailConfig *config.MailConfig) MailClient {
	ms := &mailClient{mailConfig: mailConfig}

	ms.loadTemplate()

	if ms.mailConfig.MailTemplateReloadInterval > 0 {
		go ms.startReloadLoop()
	}

	return ms
}

func (mc *mailClient) SendEmail(data *MailData) {
	go func() {
		defer mc.cleanUp(data)

		message := gomail.NewMessage()
		message.SetHeader("From", data.From)
		message.SetHeader("To", data.Recipients...)

		if data.ReplyTo != "" {
			message.SetHeader("Reply-To", data.ReplyTo)
		}
		if len(data.Cc) > 0 {
			message.SetHeader("Cc", data.Cc...)
		}

		message.SetHeader("Subject", data.Subject)

		body, err := mc.format(data)
		if err != nil {
			slog.Error("Template formatting failed", "error", err)
			return
		}
		message.SetBody("text/html", body)

		for name, path := range data.Attachments {
			message.Attach(path, gomail.Rename(name))
		}

		dialer := gomail.NewDialer(mc.mailConfig.Host, mc.mailConfig.Port, mc.mailConfig.User, mc.mailConfig.Password)
		dialer.SSL = mc.mailConfig.TlsEnabled

		if !mc.mailConfig.AuthEnabled {
			dialer.Username = ""
			dialer.Password = ""
		}

		if err := dialer.DialAndSend(message); err != nil {
			slog.Error("Email send failed", "error", err)
		}
	}()
}

func (mc *mailClient) cleanUp(data *MailData) {
	for _, path := range data.Attachments {
		if err := os.Remove(path); err != nil {
			slog.Error("Failed to delete attachment", "error", err)
		}
	}
}

func (mc *mailClient) format(data *MailData) (string, error) {
	mc.mu.RLock()
	tmpl := mc.mailTemplate
	mc.mu.RUnlock()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data.Content); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (mc *mailClient) loadTemplate() {
	const defaultTpl = `
		<h1>{{.Title}}</h1>
		{{range .Lines}}<p>{{.}}</p>{{end}}
		{{if .MailLink}}<a href="{{.MailLink.Href}}">{{.MailLink.Text}}</a>{{end}}
	`

	var tplContent string
	if mc.mailConfig.MailTemplateUrl != "" {
		data, err := os.ReadFile(mc.mailConfig.MailTemplateUrl)
		if err != nil {
			slog.Error("Failed to load mail template from file, using default", "error", err)
			tplContent = defaultTpl
		} else {
			tplContent = string(data)
		}
	} else {
		tplContent = defaultTpl
	}

	tmpl, err := template.New("mail").Parse(tplContent)
	if err != nil {
		slog.Error("Failed to parse mail template, using default", "error", err)
		tmpl, _ = template.New("mail").Parse(defaultTpl)
	}

	// Lock before replacing
	mc.mu.Lock()
	mc.mailTemplate = tmpl
	mc.mu.Unlock()

	slog.Info("Mail template reloaded successfully")
}

func (mc *mailClient) startReloadLoop() {
	ticker := time.NewTicker(mc.mailConfig.MailTemplateReloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.loadTemplate()
		}
	}
}
