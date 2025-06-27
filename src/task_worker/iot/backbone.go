package iot

import (
	"github.com/lombardidaniel/tcc/worker/models"
)

type Backbone interface {
	Forward(msgs []models.RoutingMessage) ([]models.RoutingReply, error)
}
