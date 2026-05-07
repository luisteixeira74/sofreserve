package service

import "fmt"

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) Send(to, subject, body string) {
	fmt.Printf("[EMAIL] to=%s | subject=%s | body=%s\n", to, subject, body)
}