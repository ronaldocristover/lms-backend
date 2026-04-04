package job

import (
	"log"
)

type EmailJob struct {
	To      string
	Subject string
	Body    string
}

func NewEmailJob(to, subject, body string) *EmailJob {
	return &EmailJob{
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

func (j *EmailJob) Run() error {
	// In production, integrate with SMTP or email service like SendGrid, SES, etc.
	log.Printf("Sending email to %s: %s", j.To, j.Subject)
	// TODO: Implement actual email sending
	return nil
}
