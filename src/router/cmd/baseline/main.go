package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"
	"github.com/redis/go-redis/v9"
)

var (
	err error

	mqttManager *mqtt.ConnectionManager

	dbService services.DBService

	ackChansMu sync.Mutex
	ackChans   map[string]chan bool // MAC: chan[ACK]
)

func init() {
	ackChans = make(map[string]chan bool)

	broker, _ := url.Parse("tcp://mqtt:1883")
	clientID := "baseline_router" + uuid.NewString()

	responseTopic := "/gw/+/response"
	cfg := mqtt.ClientConfig{
		BrokerUrls:        []*url.URL{broker},
		KeepAlive:         30,
		ConnectRetryDelay: 5 * time.Second,
		OnConnectionUp: func(cm *mqtt.ConnectionManager, ca *paho.Connack) {
			fmt.Println("Connected to MQTT broker")
			cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{{
					Topic: "$share/router-baseline-group/" + responseTopic,
					QoS:   byte(1),
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
					rep, err := models.RoutingReply{}.FromMqtt(pr.Packet.Payload)
					if err != nil {
						slog.Error("could not parse message")
						return false, nil
					}

					slog.Info(fmt.Sprintf("rcvd ACK: %s", rep.DeviceMac))

					ackChansMu.Lock()
					ackChan, exists := ackChans[rep.DeviceMac]
					ackChansMu.Unlock()
					if !exists {
						slog.Error(fmt.Sprintf("mac: %s recieved before chan creation", rep.DeviceMac))
						return false, nil
					}

					ackChan <- rep.Ack
					return true, nil
				},
			},
		},
	}
	mqttManager, err = mqtt.NewConnection(context.Background(), cfg)
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

	dbService = services.NewDBService(redisClient)
}

func main() {
	http.HandleFunc("POST /route", httpHandler)
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	var msg models.RoutingMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&msg); err != nil {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received POST request: %+v\n", msg)
	if err != nil {
		fmt.Printf("Error forwarding message: %s", err)
		http.Error(w, "BadGateway", http.StatusBadGateway)
		return
	}

	ackChan := make(chan bool, 10)
	ackChansMu.Lock()
	// fmt.Printf("created chan %s\n", msg.DeviceMac)
	ackChans[msg.DeviceMac] = ackChan
	ackChansMu.Unlock()

	t, _ := dbService.GetRoute(msg.DeviceMac)
	_, err = mqttManager.Publish(r.Context(), &paho.Publish{
		QoS:     1,
		Topic:   t,
		Payload: msg.Dump(),
	})

	select {
	case ack := <-ackChan:
		if ack {
			slog.Info(fmt.Sprintf("returning ACK: %s", msg.DeviceMac))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ACK received\n"))
			return
		}

		slog.Info(fmt.Sprintf("NACK: %s", msg.DeviceMac))
		http.Error(w, "NACK received", http.StatusBadGateway)
		return

	case <-time.After(60 * time.Second):
		slog.Error(fmt.Sprintf("Timedout waiding for ack: %s", msg.DeviceMac))
		http.Error(w, "Timeout waiting for ACK", http.StatusGatewayTimeout)
		return
	}
}
