package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/omaaartamer/factory-checkin-api/internal/email"
	"github.com/omaaartamer/factory-checkin-api/internal/legacy"
	"github.com/omaaartamer/factory-checkin-api/internal/model"
	"github.com/omaaartamer/factory-checkin-api/internal/queue"
	"github.com/omaaartamer/factory-checkin-api/pkg/config"
)

type Worker struct {
	queue     queue.Queue
	emailSvc  *email.EmailService
	legacyAPI *legacy.LegacyAPIClient
	config    *config.Config
	stopChan  chan bool
}

func NewWorker(q queue.Queue, cfg *config.Config) *Worker {
	return &Worker{
		queue:     q,
		emailSvc:  email.NewEmailService(cfg),
		legacyAPI: legacy.NewLegacyAPIClient(cfg),
		config:    cfg,
		stopChan:  make(chan bool),
	}
}

func (w *Worker) Start() {
	log.Println("Background worker started - processing queue messages...")

	go func() {
		for {
			select {
			case <-w.stopChan:
				log.Println("Background worker stopped")
				return
			default:
				w.processMessages()
				time.Sleep(1 * time.Second) // Check for messages every second
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stopChan <- true
}

func (w *Worker) processMessages() {
	msg, err := w.queue.Dequeue()
	if err != nil {
		log.Printf("Error dequeuing message: %v", err)
		return
	}

	if msg == nil {
		return // No messages available
	}

	log.Printf("Processing message: %s (Type: %s, Attempt: %d)", msg.ID, msg.Type, msg.Attempts)

	var processingErr error

	switch msg.Type {
	case "labor_cost_report":
		processingErr = w.processLaborCostReport(msg)
	case "email_notification":
		processingErr = w.processEmailNotification(msg)
	default:
		processingErr = fmt.Errorf("unknown message type: %s", msg.Type)
	}

	if processingErr != nil {
		log.Printf("Failed to process message %s: %v", msg.ID, processingErr)
		w.queue.MarkFailed(msg.ID)
	} else {
		log.Printf("Successfully processed message %s", msg.ID)
		w.queue.MarkCompleted(msg.ID)
	}
}

func (w *Worker) processLaborCostReport(msg *model.QueueMessage) error {
	employeeID, ok := msg.Payload["employee_id"].(string)
	if !ok {
		return fmt.Errorf("invalid employee_id in payload")
	}

	hoursWorked, ok := msg.Payload["hours_worked"].(float64)
	if !ok {
		return fmt.Errorf("invalid hours_worked in payload")
	}

	date, ok := msg.Payload["date"].(string)
	if !ok {
		return fmt.Errorf("invalid date in payload")
	}

	return w.legacyAPI.ReportHours(employeeID, hoursWorked, date)
}

func (w *Worker) processEmailNotification(msg *model.QueueMessage) error {
	employeeID, ok := msg.Payload["employee_id"].(string)
	if !ok {
		return fmt.Errorf("invalid employee_id in payload")
	}

	hoursWorked, ok := msg.Payload["hours_worked"].(float64)
	if !ok {
		return fmt.Errorf("invalid hours_worked in payload")
	}

	date, ok := msg.Payload["date"].(string)
	if !ok {
		return fmt.Errorf("invalid date in payload")
	}

	return w.emailSvc.SendWorkedHoursEmail(employeeID, hoursWorked, date)
}
