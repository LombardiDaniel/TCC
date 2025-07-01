package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/lombardidaniel/tcc/router/pkg/models"
	"github.com/lombardidaniel/tcc/router/pkg/services"
)

var (
	mqttClient mqtt.Client

	dbService        services.DBService
	messagingService services.MessagingService

	ackChans map[string]chan bool // MAC: chan[ACK]
)

func init() {
	ackChans = make(map[string]chan bool)

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

	mqttClient.Subscribe("/gw/+/response", 1, func(c mqtt.Client, m mqtt.Message) {
		rep, err := models.RoutingReply{}.FromMqtt(m.Payload())
		if err != nil {
			slog.Error("could not parse message")
			return
		}

		ackChan, exists := ackChans[rep.DeviceMac]
		if !exists {
			return
		}

		ackChan <- rep.Ack
	})

	dbService = &services.DBServiceMock{}
	messagingService = services.NewMessagingService(nil, &mqttClient)
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

	fmt.Printf("Received message: %+v\n", msg)
	t, _ := dbService.GetRoute(msg.DeviceMac)
	err := messagingService.Forward(t, msg)
	if err != nil {
		fmt.Printf("Error forwarding message: %s", err)
		http.Error(w, "BadGateway", http.StatusBadGateway)
		return
	}

	ackChans[msg.DeviceMac] = make(chan bool)

	select {
	case ack := <-ackChans[msg.DeviceMac]:
		if ack {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ACK received\n"))
		} else {
			http.Error(w, "NACK received", http.StatusBadGateway)
		}
	case <-time.After(30 * time.Second):
		http.Error(w, "Timeout waiting for ACK", http.StatusGatewayTimeout)
	}
}
