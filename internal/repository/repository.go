package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/omaaartamer/factory-checkin-api/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(databaseURL string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &Repository{db: db}

	// Create tables if they don't exist
	if err := repo.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return repo, nil
}

func (r *Repository) createTables() error {
	// Create checkin_events table
	createEventsTable := `
	CREATE TABLE IF NOT EXISTS checkin_events (
		id SERIAL PRIMARY KEY,
		employee_id VARCHAR(50) NOT NULL,
		event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('checkin', 'checkout')),
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	// Create work_sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS work_sessions (
		id SERIAL PRIMARY KEY,
		employee_id VARCHAR(50) NOT NULL,
		checkin_time TIMESTAMP WITH TIME ZONE NOT NULL,
		checkout_time TIMESTAMP WITH TIME ZONE,
		hours_worked DECIMAL(5,2),
		status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed')),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	// Create indexes
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_employee_events ON checkin_events(employee_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_employee_sessions ON work_sessions(employee_id, status);
	`

	if _, err := r.db.Exec(createEventsTable); err != nil {
		return err
	}
	if _, err := r.db.Exec(createSessionsTable); err != nil {
		return err
	}
	if _, err := r.db.Exec(createIndexes); err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateCheckinEvent(event *model.CheckinEvent) error {
	query := `
		INSERT INTO checkin_events (employee_id, event_type, timestamp)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return r.db.QueryRow(query, event.EmployeeID, event.EventType, event.Timestamp).
		Scan(&event.ID, &event.CreatedAt)
}

func (r *Repository) GetActiveSession(employeeID string) (*model.WorkSession, error) {
	var session model.WorkSession
	query := `
		SELECT id, employee_id, checkin_time, checkout_time, hours_worked, status, created_at, updated_at
		FROM work_sessions 
		WHERE employee_id = $1 AND status = 'active'
		ORDER BY created_at DESC 
		LIMIT 1`

	err := r.db.Get(&session, query, employeeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &session, err
}

func (r *Repository) CreateWorkSession(session *model.WorkSession) error {
	query := `
		INSERT INTO work_sessions (employee_id, checkin_time, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, session.EmployeeID, session.CheckinTime, session.Status).
		Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
}

func (r *Repository) UpdateWorkSession(session *model.WorkSession) error {
	query := `
		UPDATE work_sessions 
		SET checkout_time = $1, hours_worked = $2, status = $3, updated_at = NOW()
		WHERE id = $4`

	_, err := r.db.Exec(query, session.CheckoutTime, session.HoursWorked, session.Status, session.ID)
	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
