package iot

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/lombardidaniel/tcc/worker/models"
)

type BackboneRestImpl struct {
	client *http.Client
	url    string
}

func NewBackboneRestImpl(url string) Backbone {
	return &BackboneRestImpl{
		client: &http.Client{},
		url:    url,
	}
}

func (b *BackboneRestImpl) send(msg models.RoutingMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := b.client.Post(b.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("nacked")
	}

	return nil
}

func (b *BackboneRestImpl) Forward(msgs []models.RoutingMessage) ([]models.RoutingReply, error) {
	var reps []models.RoutingReply
	var repsMu sync.Mutex

	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, v := range msgs {
		wg.Add(1)
		go func(v models.RoutingMessage) {
			defer wg.Done()
			r := models.RoutingReply{DeviceMac: v.DeviceMac, Ack: false}
			err := b.send(v)
			if err != nil {
				errChan <- err
			}

			r.Ack = true

			repsMu.Lock()
			reps = append(reps, r)
			repsMu.Unlock()
		}(v)
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		return nil, err
	}

	return reps, nil
}
