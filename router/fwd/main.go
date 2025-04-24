package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var (
	mqttClient mqtt.Client
	msgs       <-chan amqp.Delivery

	messagingService services.MessagingService
	sharedMemService services.SharedMemoryService
	dbService        services.DBService
)

func init() {
	broker := "tcp://broker.hivemq.com:1883" // Replace with your broker URL
	clientID := "fwd"

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Error connecting to broker: %v\n", token.Error())
		os.Exit(1)
	}
	fmt.Println("Connected to MQTT broker")

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
		"routing", // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	s := redisClient.Ping(context.Background())
	if err := s.Err(); err != nil {
		panic(err)
	}

	messagingService = services.NewMessagingService(ch, mqttClient)
	sharedMemService = services.NewSharedMemoryService(redisClient)
	dbService = &services.DBServiceMock{}
}

func main() {
	for d := range msgs {
		go func() {
			var m models.RoutingMessage
			err := json.Unmarshal(d.Body, &m)
			if err != nil {
				log.Printf("Error unmarshalling task: %s", err)
			}
			err = sharedMemService.Save(m.DeviceMac, d.ReplyTo)
			if err != nil {
				log.Printf("Error saving to shared memory: %s", err)
			}

			t, _ := dbService.GetRoute(m.DeviceMac)
			err = messagingService.Forward(t, m)
			if err != nil {
				fmt.Printf("Error forwarding message: %s", err)
			}
		}()
	}
}
