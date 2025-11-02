package service

import (
	"fmt"
	"time"

	"github.com/omaaartamer/factory-checkin-api/internal/model"
	"github.com/omaaartamer/factory-checkin-api/internal/queue"
	"github.com/omaaartamer/factory-checkin-api/internal/repository"
)

type CheckinService struct {
	repo  *repository.Repository
	queue queue.Queue
}

func NewCheckinService(repo *repository.Repository, q queue.Queue) *CheckinService {
	return &CheckinService{
		repo:  repo,
		queue: q,
	}
}

func (s *CheckinService) ProcessCheckin(employeeID string) (*model.CheckinResponse, error) {
	now := time.Now()

	// Check if employee has an active session
	activeSession, err := s.repo.GetActiveSession(employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}

	if activeSession != nil {
		// Employee is checking out
		return s.processCheckout(employeeID, activeSession, now)
	} else {
		// Employee is checking in
		return s.processCheckin(employeeID, now)
	}
}

func (s *CheckinService) processCheckin(employeeID string, timestamp time.Time) (*model.CheckinResponse, error) {
	// Create checkin event
	event := &model.CheckinEvent{
		EmployeeID: employeeID,
		EventType:  "checkin",
		Timestamp:  timestamp,
	}

	if err := s.repo.CreateCheckinEvent(event); err != nil {
		return nil, fmt.Errorf("failed to create checkin event: %w", err)
	}

	// Create new work session
	session := &model.WorkSession{
		EmployeeID:  employeeID,
		CheckinTime: timestamp,
		Status:      "active",
	}

	if err := s.repo.CreateWorkSession(session); err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	return &model.CheckinResponse{
		Success:   true,
		Message:   "Successfully checked in",
		EventType: "checkin",
		Timestamp: timestamp,
	}, nil
}

func (s *CheckinService) processCheckout(employeeID string, activeSession *model.WorkSession, timestamp time.Time) (*model.CheckinResponse, error) {
	// Create checkout event
	event := &model.CheckinEvent{
		EmployeeID: employeeID,
		EventType:  "checkout",
		Timestamp:  timestamp,
	}

	if err := s.repo.CreateCheckinEvent(event); err != nil {
		return nil, fmt.Errorf("failed to create checkout event: %w", err)
	}

	// Calculate hours worked
	duration := timestamp.Sub(activeSession.CheckinTime)
	hoursWorked := duration.Hours()

	// Update work session
	activeSession.CheckoutTime = &timestamp
	activeSession.HoursWorked = &hoursWorked
	activeSession.Status = "completed"

	if err := s.repo.UpdateWorkSession(activeSession); err != nil {
		return nil, fmt.Errorf("failed to update work session: %w", err)
	}

	// Queue async tasks - these should not fail the checkout process
	s.queueAsyncTasks(employeeID, hoursWorked, timestamp)

	return &model.CheckinResponse{
		Success:     true,
		Message:     "Successfully checked out",
		EventType:   "checkout",
		Timestamp:   timestamp,
		HoursWorked: &hoursWorked,
	}, nil
}

func (s *CheckinService) queueAsyncTasks(employeeID string, hoursWorked float64, timestamp time.Time) {
	dateStr := timestamp.Format("2006-01-02")

	// Queue labor cost report - this is critical business data
	laborCostMsg := queue.CreateLaborCostMessage(employeeID, hoursWorked, dateStr)
	if err := s.queue.Enqueue(laborCostMsg); err != nil {
		// Log error but don't fail checkout
		fmt.Printf("WARNING: Failed to queue labor cost report for employee %s: %v\n", employeeID, err)
	}

	// Queue email notification - this is nice-to-have
	emailMsg := queue.CreateEmailMessage(employeeID, hoursWorked, dateStr)
	if err := s.queue.Enqueue(emailMsg); err != nil {
		// Log error but don't fail checkout
		fmt.Printf("WARNING: Failed to queue email notification for employee %s: %v\n", employeeID, err)
	}
}

func (s *CheckinService) GetEmployeeStatus(employeeID string) (*model.WorkSession, error) {
	return s.repo.GetActiveSession(employeeID)
}

func (s *CheckinService) GetQueueStatus() map[string]interface{} {
	return map[string]interface{}{
		"pending_messages": s.queue.GetPendingCount(),
		"timestamp":        time.Now(),
	}
}
