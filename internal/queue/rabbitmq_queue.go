package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/omaaartamer/factory-checkin-api/internal/model"
	"github.com/streadway/amqp"
)

// Queue interface - same as before for compatibility
type Queue interface {
	Enqueue(msg *model.QueueMessage) error
	Dequeue() (*model.QueueMessage, error)
	MarkCompleted(messageID string) error
	MarkFailed(messageID string) error
	GetPendingCount() int
	Close() error
}

type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQQueue(rabbitMQURL string) (*RabbitMQQueue, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		"factory_checkin_tasks", // queue name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQQueue{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (q *RabbitMQQueue) Enqueue(msg *model.QueueMessage) error {
	// Set defaults
	if msg.Status == "" {
		msg.Status = "pending"
	}
	if msg.MaxAttempts == 0 {
		msg.MaxAttempts = 5
	}

	// Serialize message
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to queue
	err = q.channel.Publish(
		"",           // exchange
		q.queue.Name, // routing key (queue name)
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // make message persistent
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("📤 Enqueued message: %s (Type: %s)", msg.ID, msg.Type)
	return nil
}

func (q *RabbitMQQueue) Dequeue() (*model.QueueMessage, error) {
	// Get one message
	delivery, ok, err := q.channel.Get(q.queue.Name, true) // auto-ack
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if !ok {
		return nil, nil // No message available
	}

	// Deserialize message
	var msg model.QueueMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	msg.Status = "processing"
	msg.Attempts++

	log.Printf("Dequeued message: %s (Type: %s, Attempt: %d)", msg.ID, msg.Type, msg.Attempts)
	return &msg, nil
}

func (q *RabbitMQQueue) MarkCompleted(messageID string) error {
	log.Printf("Message completed: %s", messageID)
	// With auto-ack, message is already removed from queue
	return nil
}

func (q *RabbitMQQueue) MarkFailed(messageID string) error {
	log.Printf("Message failed: %s", messageID)
	// In a production system, you'd implement dead letter queues here
	// For now, we'll just log the failure
	return nil
}

func (q *RabbitMQQueue) GetPendingCount() int {
	// Inspect queue to get message count
	info, err := q.channel.QueueInspect(q.queue.Name)
	if err != nil {
		log.Printf("Failed to inspect queue: %v", err)
		return 0
	}
	return info.Messages
}

func (q *RabbitMQQueue) Close() error {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		q.conn.Close()
	}
	return nil
}

// Helper functions - same as before for compatibility
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
