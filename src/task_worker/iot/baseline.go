package iot

import (
	"github.com/lombardidaniel/tcc/worker/models"
)

type BackboneRestImpl struct {
	// client http.Client
}

func NewBackboneRestImpl(tgtAddr string) Backbone {
	return &BackboneRbmqImpl{
		//
	}
}

func (b *BackboneRestImpl) Forward(msgs []models.RoutingMessage) ([]models.RoutingReply, error) {
	panic("not impl")
}
