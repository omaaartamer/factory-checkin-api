package model

import (
	"time"
)

// CheckinEvent represents a check-in or check-out event
type CheckinEvent struct {
	ID         int       `json:"id" db:"id"`
	EmployeeID string    `json:"employee_id" db:"employee_id"`
	EventType  string    `json:"event_type" db:"event_type"` // "checkin" or "checkout"
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// WorkSession represents a complete work session (checkin + checkout)
type WorkSession struct {
	ID           int        `json:"id" db:"id"`
	EmployeeID   string     `json:"employee_id" db:"employee_id"`
	CheckinTime  time.Time  `json:"checkin_time" db:"checkin_time"`
	CheckoutTime *time.Time `json:"checkout_time,omitempty" db:"checkout_time"`
	HoursWorked  *float64   `json:"hours_worked,omitempty" db:"hours_worked"`
	Status       string     `json:"status" db:"status"` // "active" or "completed"
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// QueueMessage represents a message in our processing queue
type QueueMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "labor_cost_report" or "email_notification"
	Payload     map[string]interface{} `json:"payload"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessAt   time.Time              `json:"process_at"`
	Status      string                 `json:"status"` // "pending", "processing", "completed", "failed"
}

// CheckinRequest represents the API request for check-in/check-out
type CheckinRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

// CheckinResponse represents the API response
type CheckinResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	EventType   string    `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	HoursWorked *float64  `json:"hours_worked,omitempty"`
}

// LaborCostReport represents the data sent to legacy system
type LaborCostReport struct {
	EmployeeID  string  `json:"employee_id"`
	HoursWorked float64 `json:"hours_worked"`
	Date        string  `json:"date"`
}

// EmailNotification represents email data
type EmailNotification struct {
	EmployeeID  string  `json:"employee_id"`
	HoursWorked float64 `json:"hours_worked"`
	Date        string  `json:"date"`
}
