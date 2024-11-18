package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/CoralCoralCoralCoral/simulation-engine/messaging"
	"github.com/CoralCoralCoralCoral/simulation-engine/model"
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

func NewEventSubscriber(conn *amqp091.Connection, sim_id uuid.UUID) func(event *logger.Event) {
	metrics := new(Metrics)

	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare("game-updates", "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return func(event *logger.Event) {
		metrics.applyEvent(event, ch, sim_id)
	}
}

func (metrics *Metrics) applyEvent(event *logger.Event, ch *amqp091.Channel, sim_id uuid.UUID) {
	switch event.Type {
	case model.EpochEnd:
		if payload, ok := event.Payload.(model.EpochEndPayload); ok {
			if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
				return
			}

			body, err := json.Marshal(messaging.Message{
				ApiId:       uuid.Max,
				SimServerId: uuid.Max,
				SimId:       sim_id,
				Payload:     metrics,
			})
			failOnError(err, "failed to serialize event")

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

			metrics.print(payload.Time.Format("02-01-2006"))
			metrics.reset()
		}
	case model.AgentStateUpdate:
		if payload, ok := event.Payload.(model.AgentStateUpdatePayload); ok {
			switch payload.State {
			case model.Infected:
				metrics.NewInfections += 1
				metrics.InfectedPopulation += 1
			case model.Infectious:
				metrics.InfectiousPopulation += 1
			case model.Immune:
				metrics.ImmunePopulation += 1
				metrics.NewRecoveries += 1
				metrics.InfectedPopulation -= 1
				metrics.InfectiousPopulation -= 1
			case model.Susceptible:
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
