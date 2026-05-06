// Package notifier provides alerting functionality for cronwatch.
// It sends notifications when monitored cron jobs fail or time out.
package notifier

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"text/template"
	"time"
)

// Config holds the notifier configuration.
type Config struct {
	SMTPHost string
	SMTPPort int
	From     string
	To       []string
	Username string
	Password string
}

// Alert represents a failure event to be reported.
type Alert struct {
	JobName   string
	Command   string
	ExitCode  int
	Output    string
	Err       error
	OccuredAt time.Time
}

// Notifier sends alerts via email.
type Notifier struct {
	cfg Config
}

const emailTmpl = `Subject: [cronwatch] Job "{{.JobName}}" failed
From: {{.From}}
To: {{.To}}
Content-Type: text/plain

Job:     {{.JobName}}
Command: {{.Command}}
Time:    {{.OccuredAt.Format "2006-01-02 15:04:05"}}
Exit:    {{.ExitCode}}

Output:
{{.Output}}

Error: {{.ErrStr}}
`

type emailData struct {
	Alert
	From   string
	To     string
	ErrStr string
}

// New creates a new Notifier with the given config.
func New(cfg Config) *Notifier {
	return &Notifier{cfg: cfg}
}

// Send dispatches an alert email for the given failure.
func (n *Notifier) Send(a Alert) error {
	tmpl, err := template.New("email").Parse(emailTmpl)
	if err != nil {
		return fmt.Errorf("notifier: parse template: %w", err)
	}

	errStr := ""
	if a.Err != nil {
		errStr = a.Err.Error()
	}

	data := emailData{
		Alert:  a,
		From:   n.cfg.From,
		To:     fmt.Sprintf("%v", n.cfg.To),
		ErrStr: errStr,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("notifier: render template: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", n.cfg.SMTPHost, n.cfg.SMTPPort)
	var auth smtp.Auth
	if n.cfg.Username != "" {
		auth = smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.SMTPHost)
	}

	if err := smtp.SendMail(addr, auth, n.cfg.From, n.cfg.To, buf.Bytes()); err != nil {
		return fmt.Errorf("notifier: send mail: %w", err)
	}

	log.Printf("notifier: alert sent for job %q", a.JobName)
	return nil
}
