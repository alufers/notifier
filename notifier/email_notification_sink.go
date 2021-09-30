package notifier

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type EmailNotificationSink struct {
	From         string
	To           []string
	SMTPAddress  string
	SMTPUsername string
	SMTPPassword string
	StartTLS     bool
}

func (sink *EmailNotificationSink) Init() error {
	conn, err := smtp.Dial(sink.SMTPAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server %v: %w", sink.SMTPAddress, err)
	}
	defer conn.Close()
	if sink.StartTLS {
		if err := conn.StartTLS(&tls.Config{
			ServerName: strings.Split(sink.SMTPAddress, ":")[0],
		}); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}
	err = conn.Auth(sink.getAuth())
	if err != nil {
		return fmt.Errorf("failed to authenticate to SMTP server %v: %w", sink.SMTPAddress, err)
	}
	log.Printf("Successfully initialized %T", *sink)
	return nil
}

func (sink *EmailNotificationSink) DeliverNotification(notification *Notification) error {
	errors := []error{}
	for _, to := range sink.To {

		body := fmt.Sprintf("%v\n\n\n%v", notification.Body, formatDate(notification.Timestamp))
		err := smtp.SendMail(
			sink.SMTPAddress,
			sink.getAuth(),
			sink.From,
			[]string{to},
			[]byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", sink.From, to, notification.Title, body)),
		)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		contents := []string{}
		for _, err := range errors {
			contents = append(contents, err.Error())
		}
		return fmt.Errorf(strings.Join(contents, ",\n"))
	}
	return nil
}

func (sink *EmailNotificationSink) getAuth() smtp.Auth {
	return smtp.PlainAuth("", sink.SMTPUsername, sink.SMTPPassword, strings.Split(sink.SMTPAddress, ":")[0])
}
