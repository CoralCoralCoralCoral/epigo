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

const GAME_METRICS_EXCHANGE = "game-metrics"

type MetricsTx struct {
	api_id uuid.UUID
	sim_id uuid.UUID
	ch     *amqp091.Channel
}

type Metrics struct {
	NewInfections          int
	NewHospitalizations    int
	NewRecoveries          int
	NewDeaths              int
	InfectedPopulation     int
	InfectiousPopulation   int
	HospitalizedPopulation int
	ImmunePopulation       int
	DeadPopulation         int
}

func NewMetricsTx(conn *amqp091.Connection, api_id, sim_id uuid.UUID) *MetricsTx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare(GAME_METRICS_EXCHANGE, "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return &MetricsTx{
		api_id,
		sim_id,
		ch,
	}
}

func (tx *MetricsTx) NewEventSubscriber() func(event *logger.Event) {
	metrics := new(Metrics)

	return func(event *logger.Event) {
		switch event.Type {
		case model.EpochEnd:
			if payload, ok := event.Payload.(model.EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}

				tx.send(metrics)
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
					if payload.PreviousState == model.Hospitalized {
						metrics.HospitalizedPopulation -= 1
					}
				case model.Susceptible:
					if payload.PreviousState == model.Immune {
						metrics.ImmunePopulation -= 1
					}
					if payload.PreviousState == model.Hospitalized {
						metrics.HospitalizedPopulation -= 1
					}
				case model.Hospitalized:
					metrics.NewHospitalizations += 1
					metrics.HospitalizedPopulation += 1
				case model.Dead:
					metrics.NewDeaths += 1
					metrics.DeadPopulation += 1
					if payload.PreviousState == model.Hospitalized {
						metrics.HospitalizedPopulation -= 1
					}
					if payload.HasInfectionProfile {
						metrics.InfectiousPopulation -= 1
						metrics.InfectedPopulation -= 1
					}
				default:
					panic("this should not be possible")
				}
			}
		default:
			// ignore other types of events
		}
	}
}

func (tx *MetricsTx) Close() {
	tx.ch.Close()
}

func (metrics *Metrics) reset() {
	metrics.NewInfections = 0
	metrics.NewHospitalizations = 0
	metrics.NewRecoveries = 0
	metrics.NewDeaths = 0
}

func (metrics *Metrics) print(date string) {
	fmt.Print("\033[H\033[2J")

	fmt.Printf("Epidemic state on %s\n", date)
	fmt.Printf("	New infections:				%d\n", metrics.NewInfections)
	fmt.Printf("	New hospitalizations:			%d\n", metrics.NewHospitalizations)
	fmt.Printf("	New recoveries:				%d\n", metrics.NewRecoveries)
	fmt.Printf("	New deaths:				%d\n", metrics.NewDeaths)
	fmt.Printf("	Infected population:			%d\n", metrics.InfectedPopulation)
	fmt.Printf("	Infectious population:			%d\n", metrics.InfectiousPopulation)
	fmt.Printf("	Hospitalized population:		%d\n", metrics.HospitalizedPopulation)
	fmt.Printf("	Dead population:			%d\n", metrics.DeadPopulation)
	fmt.Printf("	Immune population:			%d\n", metrics.ImmunePopulation)
}

func (tx *MetricsTx) send(metrics *Metrics) {
	routing_key := fmt.Sprintf("%s.%s", tx.api_id, tx.sim_id)

	body, err := json.Marshal(metrics)
	failOnError(err, "failed to json serialize event")

	err = tx.ch.PublishWithContext(context.Background(),
		GAME_METRICS_EXCHANGE, // exchange
		routing_key,           // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	failOnError(err, "Failed to publish message")
}
