package queue

import (
	projectConfig "github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/RichardKnop/machinery/v2"
	amqpbackend "github.com/RichardKnop/machinery/v2/backends/amqp"
	amqpbroker "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
)

func NewMachineryServer(cfg projectConfig.RabbitMQConfig, appCfg *projectConfig.AppCfg) (*machinery.Server, error) {
	cnf := &config.Config{
		Broker:        cfg.URL(),
		DefaultQueue:  "machinery_tasks",
		ResultBackend: "amqp",
		AMQP: &config.AMQPConfig{
			Exchange:     "machinery_exchange",
			ExchangeType: "direct",
			BindingKey:   "machinery_routing_key",
		},
	}

	broker := amqpbroker.New(cnf)
	backend := amqpbackend.New(cnf)
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	// Register tasks
	tasksMap := map[string]interface{}{
		"send_email": func(payloadStr string) error {
			return SendEmail(payloadStr, appCfg)
		},
		"process_payment": ProcessPayment,
	}

	return server, server.RegisterTasks(tasksMap)
}

func StartWorker(server *machinery.Server, workerTag string) error {
	worker := server.NewWorker(workerTag, 10) // 10 concurrent workers
	return worker.Launch()
}
