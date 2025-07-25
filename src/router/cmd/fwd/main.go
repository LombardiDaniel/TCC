package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
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
	broker := "tcp://mqtt:1883" // Replace with your broker URL
	clientID := "fwd" + uuid.NewString()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		fmt.Printf("Error connecting to broker: %v\n", token.Error())
		os.Exit(1)
	}
	fmt.Println("Connected to MQTT broker")

	conn, err := amqp.Dial("amqp://guest:guest@rbmq:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

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

	messagingService = services.NewMessagingService(ch, &mqttClient)
	sharedMemService = services.NewSharedMemoryService(redisClient)
	dbService = services.NewDBService(redisClient)
}

func main() {
	log.Println("Starting FWD router...")
	for d := range msgs {
		go func() {
			ctx := context.Background()
			var m models.RoutingMessage
			err := json.Unmarshal(d.Body, &m)
			if err != nil {
				log.Printf("Error unmarshalling task: %s", err)
				return
			}

			err = sharedMemService.Save(ctx, m.DeviceMac, d.ReplyTo)
			if err != nil {
				log.Printf("Error saving to shared memory: %s", err)
				return
			}

			fmt.Printf("Saved to shared memory: %s: %s", m.DeviceMac, d.ReplyTo)

			t, _ := dbService.GetRoute(m.DeviceMac)
			err = messagingService.Forward(t, m)
			if err != nil {
				fmt.Printf("Error forwarding message: %s", err)
				return
			}

			fmt.Printf("Message forwarded to topic: %s: %s", t, m.DeviceMac)
		}()
	}
}
