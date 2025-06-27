package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"

	mqtt "github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var (
	err error

	mqttManager *mqtt.ConnectionManager

	messagingService services.MessagingService
	sharedMemService services.SharedMemoryService
)

func onMsgCallback(payload []byte) error {
	log.Println("msg rcvd")
	rep, err := models.RoutingReply{}.FromMqtt(payload)
	if err != nil {
		return err
	}
	ctx := context.Background()
	// fmt.Printf("* [%s] %v\n", msg.Topic(), rep)
	k, err := sharedMemService.RepKey(ctx, rep.DeviceMac)
	if err != nil {
		return err
	}

	// log.Printf("replying to: %s\n", k)
	err = messagingService.Reply(k, rep)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	broker, _ := url.Parse("tcp://mqtt:1883")
	clientID := "bck" + uuid.NewString()

	responseTopic := "/gw/+/response"
	cfg := mqtt.ClientConfig{
		BrokerUrls:        []*url.URL{broker},
		KeepAlive:         30,
		ConnectRetryDelay: 5 * time.Second,
		OnConnectionUp: func(cm *mqtt.ConnectionManager, ca *paho.Connack) {
			fmt.Println("Connected to MQTT broker")
			cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{{
					Topic: "$share/router-bck-group/" + responseTopic,
					QoS:   byte(0),
				}},
			})
		},
		OnConnectError: func(err error) {
			fmt.Printf("Error connecting to broker: %v\n", err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					return true, onMsgCallback(pr.Packet.Payload)
				},
			},
		},
	}
	mqttManager, err = mqtt.NewConnection(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

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

	messagingService = services.NewMessagingService(ch, nil)
	sharedMemService = services.NewSharedMemoryService(redisClient)
}

func main() {
	log.Print("docker run -ti --network tcc_default eclipse-mosquitto:1.6.15 ash")
	log.Print("Example rep: `mosquitto_pub -h mqtt -t /gw/GW_MAC/response -m {\\\"deviceMac\\\":\\\"000000000001\\\",\\\"ack\\\":true}`")

	log.Println("Starting BCK router...")
	forever := make(chan struct{})
	<-forever
}
