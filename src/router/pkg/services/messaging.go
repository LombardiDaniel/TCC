package services

import (
	"github.com/lombardidaniel/tcc/router/pkg/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	// mqtt "github.com/eclipse/paho.golang/autopaho"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MessagingService is an interface that defines methods for forwarding messages
// to MQTT topics and replying to RabbitMQ queues.
type MessagingService interface {

	// Forward forwards a message to the specified topic in MQTT.
	Forward(topic string, msg models.RoutingMessage) error

	// Reply sends a reply message to the specified routing key in RabbitMQ.
	Reply(routingKey string, msg models.RoutingReply) error
}

type MessagingServiceImpl struct {
	ch         *amqp.Channel
	mqttClient *mqtt.Client
}

func NewMessagingService(ch *amqp.Channel, mqttClient *mqtt.Client) MessagingService {
	return &MessagingServiceImpl{
		ch:         ch,
		mqttClient: mqttClient,
	}
}

func (s *MessagingServiceImpl) Forward(topic string, msg models.RoutingMessage) error {
	token := (*s.mqttClient).Publish(
		topic,
		1,
		true,
		msg.Dump(),
	)
	token.Wait()
	return token.Error()
}

func (s *MessagingServiceImpl) Reply(routingKey string, msg models.RoutingReply) error {
	q, err := s.ch.QueueDeclarePassive(
		routingKey, // name
		false,      // durable
		true,       // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return err
	}

	return s.ch.Publish(
		"", // exchange
		q.Name,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body: msg.Dump(),
		},
	)
}
