package messaging

import (
	"context"
	"encoding/json"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func NewEventTransmitter(server_id uuid.UUID, sim_id uuid.UUID, conn *amqp091.Connection) func(event *logger.Event) {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare("game-updates", "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return func(event *logger.Event) {
		body, err := json.Marshal(Message{
			ApiId:       uuid.Max,
			SimServerId: server_id,
			SimId:       sim_id,
			Payload:     event,
		})
		failOnError(err, "failed to serialize evnet")

		err = ch.PublishWithContext(context.Background(),
			"game-updates", // exchange
			"test",         // routing key
			false,          // mandatory
			false,          // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			})

		failOnError(err, "Failed to publish a message")

		println(string(body))
	}
}
