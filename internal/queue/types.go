package queue

import "github.com/gofiber/fiber/v2"

type EmailPayload struct {
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	View    string            `json:"view"`
	Data    fiber.Map         `json:"data"`
}
