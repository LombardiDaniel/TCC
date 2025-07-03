package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/lombardidaniel/tcc/worker/common"
	"github.com/lombardidaniel/tcc/worker/daemons"
	"github.com/lombardidaniel/tcc/worker/domain"
	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	msgs <-chan amqp.Delivery

	dbService        services.DBService
	messagingService services.MessagingService
	telemetryService services.TelemetryService

	iotBackbone iot.Backbone

	executor domain.Executor

	taskRunner daemons.TaskRunner
)

func init() {
	ctx := context.Background()
	conn, err := amqp.Dial("amqp://guest:guest@rbmq:5672/")
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

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
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	messagingService = services.NewMessagingServiceImpl(ch)
	dbService = &services.DBServiceMock{}

	if strings.ToLower(common.GetEnvVarDefault("CUSTOM_BACKBONE", "true")) == "true" {
		iotBackbone = iot.NewBackboneImpl(
			messagingService,
		)
	} else {
		iotBackbone = iot.NewBackboneRestImpl(
			"http://router_baseline:8080/route",
		)
	}

	mongoConn := options.Client().ApplyURI(
		common.GetEnvVarDefault("MONGO_URI", "mongodb://localhost:27017"),
	)
	mongoClient, err := mongo.Connect(ctx, mongoConn)
	if err != nil {
		panic(errors.Join(err, errors.New("could not connect to mongodb")))
	}
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(errors.Join(err, errors.New("could not ping mongodb")))
	}
	tsIdxModel := mongo.IndexModel{
		Keys:    bson.M{"ts": 1},
		Options: options.Index(),
	}
	metricsCol := mongoClient.Database("tcc-telemetry").Collection("metrics")
	eventsCol := mongoClient.Database("tcc-telemetry").Collection("events")
	if _, err := metricsCol.Indexes().CreateOne(ctx, tsIdxModel); err != nil {
		panic(errors.Join(err, errors.New("could not create idx")))
	}
	if _, err := eventsCol.Indexes().CreateOne(ctx, tsIdxModel); err != nil {
		panic(errors.Join(err, errors.New("could not create idx")))
	}

	telemetryService = services.NewTelemetryServiceMongoAsyncImpl(mongoClient, metricsCol, eventsCol, 100)

	executor = domain.NewExecutor(iotBackbone, dbService, messagingService, telemetryService)
	taskRunner.RegisterTask(time.Second, telemetryService.Upload, 1)
}

func main() {
	taskRunner.Dispatch()
	log.Println("Starting task_worker...")
	ex := models.Task{
		Action:        "test",
		TransactionId: "1234567890",
		ProductId:     "1234567890",
		Ts:            time.Now(),
	}
	j, _ := json.Marshal(ex)
	log.Printf("Example task: %s\n", j)

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

		log.Printf("Received transaction: %s", task.TransactionId)

		err = executor.Execute(task) // business logic
		if err != nil {
			log.Printf("Could not execute transaction: %s", task.TransactionId)
			err := d.Acknowledger.Nack(d.DeliveryTag, false, false)
			if err != nil {
				panic(err)
			}
			continue
		}

		err = d.Acknowledger.Ack(d.DeliveryTag, false)
		if err != nil {
			panic(err)
		}
		log.Printf("Transaction %s executed successfully", task.TransactionId)
	}
}
