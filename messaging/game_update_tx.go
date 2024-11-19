package messaging

import (
	"context"
	"fmt"

	"github.com/CoralCoralCoralCoral/simulation-engine/protos/protos"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type GameUpdateTx struct {
	ch     *amqp091.Channel
	api_id uuid.UUID
	sim_id uuid.UUID
}

func NewGameUpdateTx(conn *amqp091.Connection, api_id, sim_id uuid.UUID) *GameUpdateTx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare("game-updates", "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	gu := new(GameUpdateTx)

	gu.ch = ch
	gu.api_id = api_id
	gu.sim_id = sim_id

	return gu
}

func (tx *GameUpdateTx) NewEventSubscriber() func(event *protos.Event) {
	update := new(protos.GameUpdate)

	return func(event *protos.Event) {
		switch event.Type {
		case protos.EventType_EpochEnd:
			if payload, ok := event.Payload.(*protos.Event_EpochEnd); ok {
				if (payload.EpochEnd.Epoch*payload.EpochEnd.TimeStep)%(6*60*60*1000) == 0 {
					tx.send(update)
					update = new(protos.GameUpdate)
				}
			}
		case protos.EventType_AgentStateUpdate:
			if payload, ok := event.Payload.(*protos.Event_AgentStateUpdate); ok {
				update.AgentStateUpdates = append(update.AgentStateUpdates, payload.AgentStateUpdate)
			}
		case protos.EventType_AgentLocationUpdate:
			if payload, ok := event.Payload.(*protos.Event_AgentLocationUpdate); ok {
				update.AgentLocationUpdates = append(update.AgentLocationUpdates, payload.AgentLocationUpdate)
			}
			// case protos.EventType_SpaceOccupancyUpdate:
			// 	if payload, ok := event.Payload.(*protos.Event_SpaceOccupancyUpdate); ok {
			// 		update.SpaceOccupancyUpdates = append(update.SpaceOccupancyUpdates, payload.SpaceOccupancyUpdate)
			// 	}

		}
	}
}

func (tx *GameUpdateTx) send(update *protos.GameUpdate) {
	routing_key := fmt.Sprintf("%s.%s", tx.api_id, tx.sim_id)

	body, err := proto.Marshal(update)
	fmt.Printf("sending %d bytes to game-updates", len(body))

	failOnError(err, "failed to serialize event")

	err = tx.ch.PublishWithContext(context.Background(),
		"game-updates", // exchange
		routing_key,    // routing key
		false,          // mandatory
		false,          // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	failOnError(err, "Failed to publish message")
}
