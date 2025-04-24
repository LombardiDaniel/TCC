package iot

import (
	"log"

	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
)

type Backbone interface {
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

	return nil
}
