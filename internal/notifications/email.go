package notifications

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"

	"github.com/zellis-rameesn/go-ecommerce/internal/config"
)

type SimpleEmail struct {
	To      string
	Subject string
	Body    string
}

type EmailNotifier struct {
	config *config.SMTPConfig
}

func NewEmailNotifier(cfg *config.SMTPConfig) *EmailNotifier {
	return &EmailNotifier{
		config: cfg,
	}
}

func (e *EmailNotifier) SendSimpleEmail(email *SimpleEmail) error {
	// cannot use this for ipv6
	// IPv6 addresses themselves contain : characters, so Go cannot determine where the host ends and the port begins.
	// addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	addr := net.JoinHostPort(e.config.Host, strconv.Itoa(e.config.Port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, e.config.Host)
	if err != nil {
		return err
	}

	defer client.Quit()

	if e.config.Username != "" && e.config.Password != "" {
		auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(e.config.From); err != nil {
		return err
	}
	if err := client.Rcpt(email.To); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s \r\n\r\n%s", e.config.From, email.To, email.Subject, email.Body)

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return err
	}
	return writer.Close()
}

func (e *EmailNotifier) SendLoginNotification(userEmail, userName string) error {
	email := &SimpleEmail{
		To:      userEmail,
		Subject: "Login Notification",
		Body: fmt.Sprintf(`Hello %s,

You have successfully logged into your account.

If this wasn't you, please contact support immediately.

Best regards,
The Shop Team`, userName),
	}

	return e.SendSimpleEmail(email)
}
