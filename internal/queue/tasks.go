package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Prince-Letsyo/task-management-api-go/config"
)

// Task functions for Machinery
// They must take simple serializable types or []byte

func SendEmail(payloadStr string, appCfg *config.AppCfg) error {
	var payload EmailPayload
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	body := appCfg.Mail.PrepareHTML(payload.View, payload.Data)
	err := appCfg.Mail.Send(payload.To, payload.Subject, body, "", appCfg.Mail.FromAddress)
	if err != nil {
		log.Printf("Error sending email via Machinery: %v", err)
		return err
	}
	return nil
}

func ProcessPayment(payloadStr string) error {
	log.Printf("Processing payment via Machinery: %s", payloadStr)
	// Placeholder for payment logic
	return nil
}
