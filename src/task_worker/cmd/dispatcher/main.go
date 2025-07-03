package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lombardidaniel/tcc/worker/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	count int64 = *flag.Int64("count", 1000, "ammount of messages to be sent to rbmq")

	taskQueue amqp.Queue
	ch        *amqp.Channel
)

func init() {
	conn, err := amqp.Dial("amqp://guest:guest@rbmq:5672/")
	if err != nil {
		panic(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		panic(err)
	}

	taskQueue, err = ch.QueueDeclare(
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
}

func main() {
	initTs := time.Now()
	sent := 0
	for i := range count {
		m := models.Task{
			Action:        "tagUpdate",
			TransactionId: fmt.Sprint(i),
			ProductId:     uuid.NewString(),
			Ts:            time.Now(),
		}

		body, err := json.Marshal(m)
		if err != nil {
			slog.Error(fmt.Sprintf("Could not Marshal task: %+v", m))
			continue
		}

		err = ch.Publish(
			"",
			taskQueue.Name,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         body,
			},
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Could not send task to rbmq: %s", err.Error()))
			break
		}
		sent++
	}
	deltaS := time.Since(initTs).Seconds()

	fmt.Printf("Sent %d tasks, took: %.2f\n", sent, deltaS)
}
