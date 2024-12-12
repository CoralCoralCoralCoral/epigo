package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type FeedbackTx struct {
	api_id uuid.UUID
	sim_id uuid.UUID
	ch     *amqp091.Channel
}

func NewFeedbackTx(conn *amqp091.Connection, api_id, sim_id uuid.UUID) *FeedbackTx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare(UPDATE_EXCHANGE, "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return &FeedbackTx{
		api_id,
		sim_id,
		ch,
	}
}

func (tx *FeedbackTx) NewEventSubscriber() func(event *logger.Event) {
	return func(event *logger.Event) {
		switch event.Type {
		case model.SimulationInitializing, model.SimulationInitialized, model.CommandProcessed:
			tx.send(event)
		default:
			// ignore other types of events
		}
	}
}

func (tx *FeedbackTx) Close() {
	tx.ch.Close()
}

func (tx *FeedbackTx) send(event *logger.Event) {
	routing_key := fmt.Sprintf("%s.%s", tx.api_id, tx.sim_id)

	body, err := json.Marshal(
		Notification{
			Type:    EventNotification,
			Payload: event,
		},
	)
	failOnError(err, "failed to json serialize event")

	err = tx.ch.PublishWithContext(context.Background(),
		UPDATE_EXCHANGE, // exchange
		routing_key,     // routing key
		false,           // mandatory
		false,           // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	failOnError(err, "Failed to publish message")
}
