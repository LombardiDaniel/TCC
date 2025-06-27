package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lombardidaniel/tcc/worker/domain"
	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	msgs <-chan amqp.Delivery

	dbService        services.DBService
	messagingService services.MessagingService

	iotBackbone iot.Backbone

	executor domain.Executor
)

func GetEnvVarDefault(name string, def string) string {
	envVar, ok := os.LookupEnv(name)
	if !ok {
		return def
	}

	return envVar
}

func init() {
	conn, err := amqp.Dial("amqp://guest:guest@rbmq:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		panic(err)
	}

	msgs, err = ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	messagingService = services.NewMessagingServiceImpl(ch)
	dbService = &services.DBServiceMock{}

	if strings.ToLower(GetEnvVarDefault("CUSTOM_BACKBONE", "true")) == "true" {
		iotBackbone = iot.NewBackboneImpl(
			messagingService,
		)
	} else {
		// iotBackbone = iot.NewBackboneBaselineImpl(
		// 	dbService,
		// 	messagingService,
		// )
	}

	executor = domain.NewExecutor(iotBackbone, dbService, messagingService)
}

func main() {
	log.Println("Starting task_worker...")
	ex := models.Task{
		Action:        "test",
		TransactionId: "1234567890",
		ProductId:     "1234567890",
		Ts:            time.Now(),
	}
	j, _ := json.Marshal(ex)
	log.Printf("Example task: %s\n", j)

	for d := range msgs {
		var task models.Task
		err := json.Unmarshal(d.Body, &task)
		if err != nil {
			log.Printf("Error unmarshalling task: %s", err)
			err = d.Acknowledger.Nack(d.DeliveryTag, false, false)
			if err != nil {
				panic(err)
			}
			continue
		}

		log.Printf("Received transaction: %s", task.TransactionId)

		err = executor.Execute(task) // business logic
		if err != nil {
			log.Printf("Could not execute transaction: %s", task.TransactionId)
			err := d.Acknowledger.Nack(d.DeliveryTag, false, false)
			if err != nil {
				panic(err)
			}
			continue
		}

		err = d.Acknowledger.Ack(d.DeliveryTag, false)
		if err != nil {
			panic(err)
		}
		log.Printf("Transaction %s executed successfully", task.TransactionId)
	}
}
