package domain

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/lombardidaniel/tcc/worker/common"
	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
)

var (
	isCustomBackbone bool   = strings.ToLower(common.GetEnvVarDefault("CUSTOM_BACKBONE", "true")) == "true"
	metricFlagStr    string = "custom_backbone"
	counter          services.Counter
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
	if !isCustomBackbone {
		metricFlagStr = "http_backbone"
	}

	counter, _ = telemetryService.GetCounter(context.Background(), "successfull_"+metricFlagStr, map[string]string{})

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
	// Wait for replies
	reps, err := e.backbone.Forward(msgs)
	if err != nil {
		return err
	}
	delta := time.Since(initTs)

	e.telemetryService.RecordMetric(context.Background(), "deltaTime_"+metricFlagStr, float64(delta.Milliseconds()), map[string]string{"unit": "ms"})
	counter.Increment(uint64(len(reps)))

	for _, rep := range reps {
		log.Printf("Received ACK from: %s\n", rep.DeviceMac)
	}

	if len(reps) != len(macs) {
		log.Printf("Expected %d replies, got %d\n", len(macs), len(reps))
		return errors.New("not all devices replied")
	}

	return nil
}
