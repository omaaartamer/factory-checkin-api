package email

import (
	"log"

	"github.com/omaaartamer/factory-checkin-api/pkg/config"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{config: cfg}
}

func (e *EmailService) SendWorkedHoursEmail(employeeID string, hoursWorked float64, date string) error {
	// Mock email sending (in real system, use SMTP)
	log.Printf("   MOCK EMAIL SENT:")
	log.Printf("   To: %s@company.com", employeeID)
	log.Printf("   Subject: Your work hours for %s", date)
	log.Printf("   Body: You worked %.2f hours today. Great job!", hoursWorked)
	log.Printf("   SMTP Config: %s:%d", e.config.SMTPHost, e.config.SMTPPort)

	return nil
}
