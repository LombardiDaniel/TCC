package domain

import (
	"errors"
	"log"

	"github.com/lombardidaniel/tcc/worker/iot"
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
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
}

func NewExecutor(backbone iot.Backbone, dbService services.DBService, messagingService services.MessagingService) Executor {
	return &executorImpl{
		backbone:         backbone,
		dbService:        dbService,
		messagingService: messagingService,
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

	// Wait for replies
	reps, err := e.messagingService.Forward(msgs)
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
