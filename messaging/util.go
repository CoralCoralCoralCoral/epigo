package messaging

import (
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func extractApiId(msg *amqp091.Delivery) (uuid.UUID, error) {
	parts := strings.Split(msg.RoutingKey, ".")
	if len(parts) < 1 {
		return uuid.Nil, errors.New("invalid routing key")
	}

	return uuid.Parse(parts[0])
}
