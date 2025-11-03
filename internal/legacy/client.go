package legacy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/omaaartamer/factory-checkin-api/internal/model"
	"github.com/omaaartamer/factory-checkin-api/pkg/config"
)

type LegacyAPIClient struct {
	config     *config.Config
	httpClient *http.Client
}

func NewLegacyAPIClient(cfg *config.Config) *LegacyAPIClient {
	return &LegacyAPIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (l *LegacyAPIClient) ReportHours(employeeID string, hoursWorked float64, date string) error {
	report := model.LaborCostReport{
		EmployeeID:  employeeID,
		HoursWorked: hoursWorked,
		Date:        date,
	}

	jsonData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Mock the legacy API call (in real system, this would call the actual legacy system)
	log.Printf("MOCK LEGACY API CALL:")
	log.Printf("URL: %s", l.config.LegacyAPIURL)
	log.Printf("Employee: %s", employeeID)
	log.Printf("Hours: %.2f", hoursWorked)
	log.Printf("Date: %s", date)
	log.Printf("Payload: %s", string(jsonData))

	// Simulate network delay
	time.Sleep(500 * time.Millisecond)

	return nil
}
