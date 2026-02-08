package queue

import (
	"encoding/json"
	"fmt"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/tasks"
)

type WorkerClient struct {
	server *machinery.Server
}

func NewWorkerClient(server *machinery.Server) *WorkerClient {
	return &WorkerClient{server: server}
}

func (c *WorkerClient) EnqueueEmail(payload EmailPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	signature := &tasks.Signature{
		Name: "send_email",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(data),
			},
		},
	}

	_, err = c.server.SendTask(signature)
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}
	return nil
}
