package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CoralCoralCoralCoral/simulation-engine/messaging"
	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/CoralCoralCoralCoral/simulation-engine/protos/protos"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Metrics struct {
	NewInfections        int
	NewRecoveries        int
	InfectedPopulation   int
	InfectiousPopulation int
	ImmunePopulation     int
}

func NewEventSubscriber(conn *amqp091.Connection, api_id, sim_id uuid.UUID) func(event *protos.Event) {
	metrics := new(Metrics)

	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare("game-metrics", "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return func(event *protos.Event) {
		metrics.applyEvent(event, ch, api_id, sim_id)
	}
}

func (metrics *Metrics) applyEvent(event *protos.Event, ch *amqp091.Channel, api_id, sim_id uuid.UUID) {
	switch event.Type {
	case protos.EventType_EpochEnd:
		if payload, ok := event.Payload.(*protos.Event_EpochEnd); ok {
			if (payload.EpochEnd.Epoch*payload.EpochEnd.TimeStep)%(24*60*60*1000) != 0 {
				return
			}

			routing_key := fmt.Sprintf("%s.%s", api_id, sim_id)

			body, err := json.Marshal(messaging.Message{
				Payload: metrics,
			})
			failOnError(err, "failed to serialize event")

			err = ch.PublishWithContext(context.Background(),
				"game-updates", // exchange
				routing_key,    // routing key
				false,          // mandatory
				false,          // immediate
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        body,
				})

			failOnError(err, "Failed to publish a message")

			metrics.print(payload.EpochEnd.GetTime().String())
			metrics.reset()
		}
	case protos.EventType_AgentStateUpdate:
		if payload, ok := event.Payload.(*protos.Event_AgentStateUpdate); ok {
			switch payload.AgentStateUpdate.State {
			case string(model.Infected):
				metrics.NewInfections += 1
				metrics.InfectedPopulation += 1
			case string(model.Infectious):
				metrics.InfectiousPopulation += 1
			case string(model.Immune):
				metrics.ImmunePopulation += 1
				metrics.NewRecoveries += 1
				metrics.InfectedPopulation -= 1
				metrics.InfectiousPopulation -= 1
			case string(model.Susceptible):
				metrics.ImmunePopulation -= 1
			default:
				panic("this should not be possible")
			}
		}
	default:
		// ignore other types of events
	}
}

func (metrics *Metrics) reset() {
	metrics.NewInfections = 0
	metrics.NewRecoveries = 0
}

func (metrics *Metrics) print(date string) {
	fmt.Print("\033[H\033[2J")

	fmt.Printf("Epidemic state on %s\n", date)
	fmt.Printf("	New infections:			%d\n", metrics.NewInfections)
	fmt.Printf("	New recoveries:			%d\n", metrics.NewRecoveries)
	fmt.Printf("	Infected population:		%d\n", metrics.InfectedPopulation)
	fmt.Printf("	Infectious population:		%d\n", metrics.InfectiousPopulation)
	fmt.Printf("	Immune population:		%d\n", metrics.ImmunePopulation)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
