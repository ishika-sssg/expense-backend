package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"gopkg.in/gomail.v2"
)

func ParseBody(r *http.Request, x interface{}) {
	if body, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal([]byte(body), x); err != nil {
			return
		}
	}
}

func SendEmail(recipientEmail string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", fmt.Sprintf("%s <%s>", config.DefaultEmailSettings.SenderName, config.DefaultEmailSettings.SenderEmail))
	mailer.SetHeader("To", recipientEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body) // Use "text/plain" for plain text emails

	dialer := gomail.NewDialer(
		config.DefaultEmailSettings.SMTPHost,
		config.DefaultEmailSettings.SMTPPort,
		config.DefaultEmailSettings.AuthEmail,
		config.DefaultEmailSettings.AuthPassword,
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
