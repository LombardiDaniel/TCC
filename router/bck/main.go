package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"

	// mqtt "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/eclipse/paho.golang/autopaho"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var (
	mqttClient mqtt.ClientConfig

	messagingService services.MessagingService
	sharedMemService services.SharedMemoryService
)

func init() {
	broker := "tcp://mqtt:1883"
	clientID := "bck"

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
	log.Print("docker run -ti --network tcc_default eclipse-mosquitto:1.6.15 ash")
	log.Print("Example rep: `mosquitto_pub -h mqtt -t /gw/GW_MAC/response -m {\\\"deviceMac\\\":\\\"000000000001\\\",\\\"ack\\\":true}`")

	responseTopic := "/gw/+/response" // topic: "/gw/ANY_GW_MAC/response"

	// "$share/GROUP_NAME/" enables sharing (round-robin) in messages for MQTT
	mqttClient.Subscribe("$share/router-bck-group/"+responseTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Println("msg rcvd")
		rep, err := models.RoutingReply{}.FromMqtt(msg.Payload())
		if err != nil {
			return
		}
		ctx := context.Background()
		// fmt.Printf("* [%s] %v\n", msg.Topic(), rep)
		k, err := sharedMemService.RepKey(ctx, rep.DeviceMac)
		if err != nil {
			return
		}

		err = messagingService.Reply(k, rep)
		if err != nil {
			return
		}
	})

	log.Println("Starting BCK router...")
	forever := make(chan struct{})
	<-forever
}
