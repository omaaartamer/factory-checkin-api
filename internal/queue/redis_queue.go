package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
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

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

func (q *RedisQueue) Enqueue(msg *model.QueueMessage) error {
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

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Store message in Redis hash
	if err := q.client.HSet(q.ctx, "messages", msg.ID, data).Err(); err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	// Add to pending queue
	if err := q.client.LPush(q.ctx, "pending", msg.ID).Err(); err != nil {
		return fmt.Errorf("failed to queue message: %w", err)
	}

	return nil
}

func (q *RedisQueue) Dequeue() (*model.QueueMessage, error) {
	// Block for up to 1 second waiting for a message
	result, err := q.client.BRPop(q.ctx, 1*time.Second, "pending").Result()
	if err == redis.Nil {
		return nil, nil // No message available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue message: %w", err)
	}

	messageID := result[1]

	// Get message data
	data, err := q.client.HGet(q.ctx, "messages", messageID).Result()
	if err == redis.Nil {
		return nil, nil // Message not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message data: %w", err)
	}

	// Deserialize message
	var msg model.QueueMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}

	// Update status and attempts
	msg.Status = "processing"
	msg.Attempts++

	// Save updated message
	updatedData, _ := json.Marshal(msg)
	q.client.HSet(q.ctx, "messages", messageID, updatedData)

	return &msg, nil
}

func (q *RedisQueue) MarkCompleted(messageID string) error {
	// Update message status
	data, err := q.client.HGet(q.ctx, "messages", messageID).Result()
	if err != nil {
		return fmt.Errorf("message not found: %s", messageID)
	}

	var msg model.QueueMessage
	json.Unmarshal([]byte(data), &msg)
	msg.Status = "completed"

	updatedData, _ := json.Marshal(msg)
	return q.client.HSet(q.ctx, "messages", messageID, updatedData).Err()
}

func (q *RedisQueue) MarkFailed(messageID string) error {
	data, err := q.client.HGet(q.ctx, "messages", messageID).Result()
	if err != nil {
		return fmt.Errorf("message not found: %s", messageID)
	}

	var msg model.QueueMessage
	json.Unmarshal([]byte(data), &msg)

	if msg.Attempts < msg.MaxAttempts {
		// Retry logic - add back to queue with delay
		msg.Status = "pending"
		retryDelay := time.Duration(msg.Attempts*msg.Attempts) * time.Minute
		msg.ProcessAt = time.Now().Add(retryDelay)

		updatedData, _ := json.Marshal(msg)
		q.client.HSet(q.ctx, "messages", messageID, updatedData)

		// Re-queue after delay (simplified - in production you'd use Redis delayed jobs)
		q.client.LPush(q.ctx, "pending", messageID)
	} else {
		// Max attempts reached
		msg.Status = "failed"
		updatedData, _ := json.Marshal(msg)
		q.client.HSet(q.ctx, "messages", messageID, updatedData)
	}

	return nil
}

func (q *RedisQueue) GetPendingCount() int {
	count, _ := q.client.LLen(q.ctx, "pending").Result()
	return int(count)
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
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
