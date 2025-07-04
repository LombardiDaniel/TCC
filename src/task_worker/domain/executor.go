package domain

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
)

var (
	experimentName string = os.Getenv("EXPERIMENT_NAME")
	routerReplicas        = os.Getenv("ROUTER_REPLICAS")
	workerReplicas        = os.Getenv("WORKER_REPLICAS")
)

type Executor interface {
	// Execute executes the task and returns an error if it fails.
	// It holds all the business logic
	Execute(task models.Task) error
}

type executorImpl struct {
	backbone         iot.Backbone
	dbService        services.DBService
	messagingService services.MessagingService
	telemetryService services.TelemetryService
}

func NewExecutor(backbone iot.Backbone, dbService services.DBService, messagingService services.MessagingService, telemetryService services.TelemetryService) Executor {
	return &executorImpl{
		backbone:         backbone,
		dbService:        dbService,
		messagingService: messagingService,
		telemetryService: telemetryService,
	}
}

func (e *executorImpl) Execute(task models.Task) error {

	macs, err := e.dbService.GetMacs(task.ProductId)
	if err != nil {
		return err
	}

	var msgs []models.RoutingMessage
	for _, mac := range macs {
		m := models.RoutingMessage{
			DeviceMac: mac,
			Type:      models.ImageRequest,
			Fields: map[string]string{
				"url": "https://example.com/image.jpg",
			},
		}
		msgs = append(msgs, m)
	}

	initTs := time.Now()
	reps, err := e.backbone.Forward(msgs) // waits for replies
	delta := time.Since(initTs)
	e.telemetryService.RecordMetric(
		context.Background(),
		"backbone-execution-deltatime",
		float64(delta.Milliseconds()),
		map[string]string{
			"unit":            "ms",
			"experiment":      experimentName,
			"router_replicas": routerReplicas,
			"worker_replicas": workerReplicas,
		},
	)

	if err != nil {
		return err
	}

	for _, rep := range reps {
		log.Printf("Received ACK from: %s\n", rep.DeviceMac)
	}

	if len(reps) != len(macs) {
		log.Printf("Expected %d replies, got %d\n", len(macs), len(reps))
		return errors.New("not all devices replied")
	}

	return nil
}
