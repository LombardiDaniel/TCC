package main

import (
	"encoding/json"
	"log"

	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	numGoroutines = 5
)

var (
	msgs <-chan amqp.Delivery

	messagingService services.MessagingService

	iotBackbone iot.Backbone
)

func init() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

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
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}
}

func consume() {
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
		err = d.Acknowledger.Ack(d.DeliveryTag, false)
		if err != nil {
			panic(err)
		}

		err = iotBackbone.Execute(task)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	for range numGoroutines {
		go consume()
	}

	forever := make(chan struct{})
	<-forever
}
