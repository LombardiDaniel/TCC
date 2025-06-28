package main

import (
	"encoding/json"
	"fmt"
	"log"
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

	// TODO: precisa colocar o reply, pegar de um chan (?)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Received\n"))
}
