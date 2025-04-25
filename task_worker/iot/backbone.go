package iot

import (
	"errors"
	"log"

	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
)

type Backbone interface {
	// Execute executes the task and returns an error if it fails.
	// It holds business logic and is has access to the MessagingService.Forward()
	// method, it handles all async logic.
	Execute(task models.Task) error
}

type BackboneImpl struct {
	dbService        services.DBService
	messagingService services.MessagingService
}

func NewBackboneImpl(dbService services.DBService, messagingService services.MessagingService) Backbone {
	return &BackboneImpl{
		dbService:        dbService,
		messagingService: messagingService,
	}
}

func (b *BackboneImpl) Execute(task models.Task) error {

	macs, err := b.dbService.GetMacs(task.ProductId)
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
	reps, err := b.messagingService.Forward(msgs)
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
