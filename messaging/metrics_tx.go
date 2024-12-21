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

type MetricsTx struct {
	api_id uuid.UUID
	sim_id uuid.UUID
	ch     *amqp091.Channel
}

type JuristictionMetrics map[string]*Metrics

type Metrics struct {
	// not a serialized field
	jurisdiction *model.Jurisdiction

	Day int `json:"day"`

	NewInfections          int `json:"new_infections"`
	NewHospitalizations    int `json:"new_hospitalizations"`
	NewRecoveries          int `json:"new_recoveries"`
	NewDeaths              int `json:"new_deaths"`
	InfectedPopulation     int `json:"infected_population"`
	InfectiousPopulation   int `json:"infectious_population"`
	HospitalizedPopulation int `json:"hospitalized_population"`
	ImmunePopulation       int `json:"immune_population"`
	DeadPopulation         int `json:"dead_population"`

	// metrics yielded by surveillance processes
	NewTests           int `json:"new_tests"`
	NewPositiveTests   int `json:"new_positive_tests"`
	TotalTests         int `json:"total_tests"`
	TotalPositiveTests int `json:"total_positive_tests"`
	TestBacklog        int `json:"test_backlog"`
	TestCapacity       int `json:"test_capacity"`
}

func NewMetricsTx(conn *amqp091.Connection, api_id, sim_id uuid.UUID) *MetricsTx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare(NOTIFICATION_EXCHANGE, "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	return &MetricsTx{
		api_id,
		sim_id,
		ch,
	}
}

func (tx *MetricsTx) NewEventSubscriber() func(event *logger.Event) {
	juristiction_metrics := make(JuristictionMetrics)

	day := 0

	return func(event *logger.Event) {
		switch event.Type {
		case model.EpochEnd:
			if payload, ok := event.Payload.(model.EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}

				day += 1
				for _, metrics := range juristiction_metrics {
					metrics.Day = day
				}

				tx.send(juristiction_metrics)
				// juristiction_metrics.print(payload.Time.Format("02-01-2006"))
				juristiction_metrics.reset()
			}
		case model.AgentStateUpdate:
			if payload, ok := event.Payload.(model.AgentStateUpdatePayload); ok {
				juristiction_metrics.applyAgentStateUpdate(payload.Jurisdiction(), &payload)
			}
		case model.SpaceTestingUpdate:
			if payload, ok := event.Payload.(model.SpaceTestingUpdatePayload); ok {
				juristiction_metrics.applySpaceTestingUpdate(payload.Jurisdiction(), &payload)
			}
		default:
			// ignore other types of events
		}
	}
}

func (tx *MetricsTx) Close() {
	tx.ch.Close()
}

func (juristiction_metrics JuristictionMetrics) applySpaceTestingUpdate(jur *model.Jurisdiction, payload *model.SpaceTestingUpdatePayload) {
	jur_id := jur.Id

	if _, ok := juristiction_metrics[jur_id]; !ok {
		juristiction_metrics[jur_id] = &Metrics{jurisdiction: jur}
	}

	metrics := juristiction_metrics[jur_id]

	metrics.TotalTests += int(payload.Negatives) + int(payload.Positives)
	metrics.TotalPositiveTests += int(payload.Positives)
	metrics.NewTests += int(payload.Negatives) + int(payload.Positives)
	metrics.NewPositiveTests += int(payload.Positives)
	metrics.TestBacklog += int(payload.Backlog)
	metrics.TestCapacity += int(payload.Capacity)

	if parent := jur.Parent(); parent != nil {
		juristiction_metrics.applySpaceTestingUpdate(parent, payload)
	}
}

func (juristiction_metrics JuristictionMetrics) applyAgentStateUpdate(jur *model.Jurisdiction, payload *model.AgentStateUpdatePayload) {
	jur_id := jur.Id

	if _, ok := juristiction_metrics[jur_id]; !ok {
		juristiction_metrics[jur_id] = &Metrics{jurisdiction: jur}
	}

	metrics := juristiction_metrics[jur_id]

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

	if parent := jur.Parent(); parent != nil {
		juristiction_metrics.applyAgentStateUpdate(parent, payload)
	}
}

func (juristiction_metrics JuristictionMetrics) reset() {
	for _, metrics := range juristiction_metrics {
		metrics.reset()
	}
}

func (juristiction_metrics JuristictionMetrics) print(date string) {
	// fmt.Print("\033[H\033[2J")

	juristiction_metrics["GLOBAL"].print(date)
}

func (metrics *Metrics) reset() {
	metrics.NewInfections = 0
	metrics.NewHospitalizations = 0
	metrics.NewRecoveries = 0
	metrics.NewDeaths = 0

	metrics.NewTests = 0
	metrics.NewPositiveTests = 0
	metrics.TestBacklog = 0  // since the backlog is reported daily, reset it
	metrics.TestCapacity = 0 // since the capacity is reported faily, reset it
}

func (metrics *Metrics) print(date string) {
	fmt.Printf("Epidemic state for %s on %s\n", metrics.jurisdiction.Id, date)
	fmt.Printf("	New infections:				%d\n", metrics.NewInfections)
	fmt.Printf("	New hospitalizations:			%d\n", metrics.NewHospitalizations)
	fmt.Printf("	New recoveries:				%d\n", metrics.NewRecoveries)
	fmt.Printf("	New deaths:				%d\n", metrics.NewDeaths)
	fmt.Printf("	Infected population:			%d\n", metrics.InfectedPopulation)
	fmt.Printf("	Infectious population:			%d\n", metrics.InfectiousPopulation)
	fmt.Printf("	Hospitalized population:		%d\n", metrics.HospitalizedPopulation)
	fmt.Printf("	Dead population:			%d\n", metrics.DeadPopulation)
	fmt.Printf("	Immune population:			%d\n", metrics.ImmunePopulation)

	// surveillance related metrics
	fmt.Print("\n")
	fmt.Printf("	New tests:				%d\n", metrics.NewTests)
	fmt.Printf("	New detected cases:			%d\n", metrics.NewPositiveTests)
	fmt.Printf("	Total Tests performed:			%d\n", metrics.TotalTests)
	fmt.Printf("	Total Detected cases:				%d\n", metrics.TotalPositiveTests)
	fmt.Printf("	Test backlog:				%d\n", metrics.TestBacklog)
	fmt.Printf("	Test capacity:				%d\n", metrics.TestCapacity)
}

func (tx *MetricsTx) send(juristiction_metrics JuristictionMetrics) {
	routing_key := fmt.Sprintf("%s.%s", tx.api_id, tx.sim_id)

	body, err := json.Marshal(Notification{
		Type:    MetricsNotification,
		Payload: juristiction_metrics,
	})
	failOnError(err, "failed to json serialize event")

	err = tx.ch.PublishWithContext(context.Background(),
		NOTIFICATION_EXCHANGE, // exchange
		routing_key,           // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	failOnError(err, "Failed to publish message")
}
