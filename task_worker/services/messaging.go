package services

import (
	"encoding/json"
	"time"

	"github.com/lombardidaniel/tcc/worker/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	replyTimeout     = 5 * 5 * time.Second
	routingQueueName = "routing"
)

type MessagingService interface {
	Forward(msg []models.RoutingMessage) ([]models.RoutingReply, error)
}

type MessagingServiceImpl struct {
	ch *amqp.Channel
}

func NewMessagingServiceImpl(ch *amqp.Channel) MessagingService {
	return &MessagingServiceImpl{
		ch: ch,
	}
}

func (s *MessagingServiceImpl) Forward(msgs []models.RoutingMessage) ([]models.RoutingReply, error) {
	// Ephemeral queue is declared
	q, err := s.ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	// All msgs Published
	for _, msg := range msgs {
		msgJson, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		err = s.ch.Publish(
			"",               // exchange
			routingQueueName, // routing key
			false,            // mandatory
			false,            // immediate
			amqp.Publishing{
				ReplyTo: q.Name,
				Body:    msgJson,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	repMsgs, err := s.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	// Wait for replies
	var reps []models.RoutingReply
	select {
	case msg := <-repMsgs:
		var replyMsg models.RoutingReply
		err = json.Unmarshal(msg.Body, &replyMsg)
		if replyMsg.Ack {
			reps = append(reps, replyMsg)
			// Early return if we have all replies
			if len(reps) == len(msgs) {
				return reps, nil
			}
		}
	case <-time.After(replyTimeout):
		return reps, nil
	}

	return reps, nil
}
