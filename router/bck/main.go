package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var (
	mqttClient mqtt.Client

	messagingService services.MessagingService
	sharedMemService services.SharedMemoryService
)

func init() {
	broker := "tcp://mqtt:1883"
	clientID := "bck"

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
}

func main() {
	responseTopic := "/gw/+/response"
	mqttClient.Subscribe(responseTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		rep, err := models.RoutingReply{}.FromMqtt(msg.Payload())
		if err != nil {
			return
		}
		// fmt.Printf("* [%s] %v\n", msg.Topic(), rep)
		k, err := sharedMemService.RepKey(rep.DeviceMac)
		if err != nil {
			return
		}

		err = messagingService.Reply(k, rep)
		if err != nil {
			return
		}
	})

	forever := make(chan struct{})
	<-forever
}
