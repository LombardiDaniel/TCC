package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/lombardidaniel/tcc/router/pkg/models"

	mqtt "github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

var (
	err error
	ctx context.Context

	mqttManager *mqtt.ConnectionManager
)

func onMsgCallback(topic string, payload []byte) error {
	routingMsg, err := models.RoutingMessage{}.FromMqtt(payload)
	if err != nil {
		log.Printf("error unmarshling the RoutingMessage request: %s", string(payload))
		return err
	}

	rep := models.RoutingReply{
		DeviceMac: routingMsg.DeviceMac,
		Ack:       true,
	}

	// fmt.Println(len(topic))
	// fmt.Println(topic)

	if len(topic) != 23 {
		log.Printf("invalid topic: %s", string(payload))
		return errors.New("unrecognized MAC")
	}

	gwMac := topic[4:16]

	_, err = mqttManager.Publish(ctx, &paho.Publish{
		QoS:     1,
		Topic:   "/gw/" + gwMac + "/response",
		Payload: rep.Dump(),
	})
	if err != nil {
		log.Printf("could not publish msg: %s", err.Error())
		return err
	}

	return nil
}

func init() {
	ctx = context.Background()
	broker, _ := url.Parse("tcp://mqtt:1883")
	clientID := "esl" + uuid.NewString()

	actionTopic := "/gw/+/action"
	cfg := mqtt.ClientConfig{
		BrokerUrls:        []*url.URL{broker},
		KeepAlive:         30,
		ConnectRetryDelay: 5 * time.Second,
		OnConnectionUp: func(cm *mqtt.ConnectionManager, ca *paho.Connack) {
			fmt.Println("Connected to MQTT broker")
			cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{{
					Topic: "$share/router-bck-group/" + actionTopic,
					QoS:   byte(1),
				}},
			})
		},
		OnConnectError: func(err error) {
			log.Panicf("Error connecting to broker: %v\n", err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					return true, onMsgCallback(pr.Packet.Topic, pr.Packet.Payload)
				},
			},
		},
	}
	mqttManager, err = mqtt.NewConnection(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
}

func main() {
	// mosquitto_pub -h mqtt -t /gw/001a2b3c4d5e/action -m {\"deviceMac\":\"001a2b3c4d5e\",\"type\":0,\"fields\":{}}
	log.Println("Starting ESL emulator...")
	forever := make(chan struct{})
	<-forever
}
