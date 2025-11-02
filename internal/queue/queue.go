package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/omaaartamer/factory-checkin-api/internal/model"
)

type Queue interface {
	Enqueue(msg *model.QueueMessage) error
	Dequeue() (*model.QueueMessage, error)
	MarkCompleted(messageID string) error
	MarkFailed(messageID string) error
	GetPendingCount() int
	Close() error
}

// InMemoryQueue implements Queue interface using in-memory storage
type InMemoryQueue struct {
	messages map[string]*model.QueueMessage
	pending  chan string
	mu       sync.RWMutex
	closed   bool
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		messages: make(map[string]*model.QueueMessage),
		pending:  make(chan string, 1000), // Buffer for 1000 messages
	}
}

func (q *InMemoryQueue) Enqueue(msg *model.QueueMessage) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	// Generate ID if not provided
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("%d_%s", time.Now().UnixNano(), msg.Type)
	}

	// Set defaults
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.ProcessAt.IsZero() {
		msg.ProcessAt = time.Now()
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	if msg.MaxAttempts == 0 {
		msg.MaxAttempts = 5
	}

	// Store message
	q.messages[msg.ID] = msg

	// Add to pending queue if ready to process
	if msg.ProcessAt.Before(time.Now()) || msg.ProcessAt.Equal(time.Now()) {
		select {
		case q.pending <- msg.ID:
			// Successfully queued
		default:
			// Queue is full, but message is stored
			return fmt.Errorf("queue is full, message stored but not queued for immediate processing")
		}
	}

	return nil
}

func (q *InMemoryQueue) Dequeue() (*model.QueueMessage, error) {
	select {
	case messageID := <-q.pending:
		q.mu.RLock()
		msg, exists := q.messages[messageID]
		q.mu.RUnlock()

		if !exists {
			// Message was deleted, try next
			return q.Dequeue()
		}

		// Update status and attempts
		q.mu.Lock()
		msg.Status = "processing"
		msg.Attempts++
		q.mu.Unlock()

		return msg, nil
	default:
		return nil, nil // No messages available
	}
}

func (q *InMemoryQueue) MarkCompleted(messageID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg, exists := q.messages[messageID]
	if !exists {
		return fmt.Errorf("message not found: %s", messageID)
	}

	msg.Status = "completed"
	// In a real implementation, you might want to move completed messages to a different store
	// For now, we'll keep them in memory for debugging
	return nil
}

func (q *InMemoryQueue) MarkFailed(messageID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg, exists := q.messages[messageID]
	if !exists {
		return fmt.Errorf("message not found: %s", messageID)
	}

	if msg.Attempts < msg.MaxAttempts {
		// Retry with exponential backoff
		msg.Status = "pending"
		retryDelay := time.Duration(msg.Attempts*msg.Attempts) * time.Minute
		msg.ProcessAt = time.Now().Add(retryDelay)

		// Re-queue for later processing
		go func() {
			time.Sleep(retryDelay)
			q.mu.Lock()
			if !q.closed && msg.Status == "pending" {
				select {
				case q.pending <- messageID:
					// Successfully re-queued
				default:
					// Queue is full, will try again later
				}
			}
			q.mu.Unlock()
		}()
	} else {
		// Max attempts reached
		msg.Status = "failed"
	}

	return nil
}

func (q *InMemoryQueue) GetPendingCount() int {
	return len(q.pending)
}

func (q *InMemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closed = true
	close(q.pending)
	return nil
}

// Helper function to create queue messages
func CreateLaborCostMessage(employeeID string, hoursWorked float64, date string) *model.QueueMessage {
	return &model.QueueMessage{
		Type: "labor_cost_report",
		Payload: map[string]interface{}{
			"employee_id":  employeeID,
			"hours_worked": hoursWorked,
			"date":         date,
		},
		MaxAttempts: 5,
	}
}

func CreateEmailMessage(employeeID string, hoursWorked float64, date string) *model.QueueMessage {
	return &model.QueueMessage{
		Type: "email_notification",
		Payload: map[string]interface{}{
			"employee_id":  employeeID,
			"hours_worked": hoursWorked,
			"date":         date,
		},
		MaxAttempts: 3,
	}
}
