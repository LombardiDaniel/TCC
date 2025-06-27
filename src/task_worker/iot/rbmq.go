package iot

import (
	"github.com/lombardidaniel/tcc/worker/models"
	"github.com/lombardidaniel/tcc/worker/services"
)

type BackboneRbmqImpl struct {
	messagingService services.MessagingService
}

func NewBackboneImpl(messagingService services.MessagingService) Backbone {
	return &BackboneRbmqImpl{
		messagingService: messagingService,
	}
}

func (b *BackboneRbmqImpl) Forward(msgs []models.RoutingMessage) ([]models.RoutingReply, error) {
	return b.messagingService.Forward(msgs)
}
