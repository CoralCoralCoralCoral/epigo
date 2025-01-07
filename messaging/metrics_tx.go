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

	// space surveillance metrics
	NewTests           int `json:"new_tests"`
	NewPositiveTests   int `json:"new_positive_tests"`
	TotalTests         int `json:"total_tests"`
	TotalPositiveTests int `json:"total_positive_tests"`
	TestBacklog        int `json:"test_backlog"`
	TestCapacity       int `json:"test_capacity"`

	// cases are yielded by space surveillance processes, but attributed
	// to the agent's home jurisdiction rather than that of the space
	// that carried out the test. Sometimes tests are carried out in
	// jurisdictions other than the patient's home jurisdiction
	NewCases   int `json:"new_cases"`
	TotalCases int `json:"total_cases"`
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
	jurisdiction_metrics := make(JuristictionMetrics)

	day := 0

	return func(event *logger.Event) {
		switch event.Type {
		case model.EpochEnd:
			if payload, ok := event.Payload.(model.EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}

				day += 1
				for _, metrics := range jurisdiction_metrics {
					metrics.Day = day
				}

				tx.send(jurisdiction_metrics)
				jurisdiction_metrics.reset()
			}
		case model.AgentStateUpdate:
			if payload, ok := event.Payload.(model.AgentStateUpdatePayload); ok {
				jurisdiction_metrics.applyAgentStateUpdate(payload.Jurisdiction(), &payload)
			}
		case model.CaseDetected:
			if payload, ok := event.Payload.(model.CaseDetectedPayload); ok {
				jurisdiction_metrics.applyCaseDetected(payload.Jurisdiction())
			}
		case model.SpaceTestingUpdate:
			if payload, ok := event.Payload.(model.SpaceTestingUpdatePayload); ok {
				jurisdiction_metrics.applySpaceTestingUpdate(payload.Jurisdiction(), &payload)
			}
		default:
			// ignore other types of events
		}
	}
}

func (tx *MetricsTx) Close() {
	tx.ch.Close()
}

func (jurisdiction_metrics JuristictionMetrics) applySpaceTestingUpdate(jur *model.Jurisdiction, payload *model.SpaceTestingUpdatePayload) {
	jur_id := jur.Id

	if _, ok := jurisdiction_metrics[jur_id]; !ok {
		jurisdiction_metrics[jur_id] = &Metrics{jurisdiction: jur}
	}

	metrics := jurisdiction_metrics[jur_id]

	metrics.TotalTests += int(payload.Negatives) + int(payload.Positives)
	metrics.TotalPositiveTests += int(payload.Positives)
	metrics.NewTests += int(payload.Negatives) + int(payload.Positives)
	metrics.NewPositiveTests += int(payload.Positives)
	metrics.TestBacklog += int(payload.Backlog)
	metrics.TestCapacity += int(payload.Capacity)

	if parent := jur.Parent(); parent != nil {
		jurisdiction_metrics.applySpaceTestingUpdate(parent, payload)
	}
}

func (jurisdiction_metrics JuristictionMetrics) applyCaseDetected(jur *model.Jurisdiction) {
	jur_id := jur.Id

	if _, ok := jurisdiction_metrics[jur_id]; !ok {
		jurisdiction_metrics[jur_id] = &Metrics{jurisdiction: jur}
	}

	metrics := jurisdiction_metrics[jur_id]

	metrics.NewCases += 1
	metrics.TotalCases += 1

	if parent := jur.Parent(); parent != nil {
		jurisdiction_metrics.applyCaseDetected(parent)
	}
}

func (jurisdiction_metrics JuristictionMetrics) applyAgentStateUpdate(jur *model.Jurisdiction, payload *model.AgentStateUpdatePayload) {
	jur_id := jur.Id

	if _, ok := jurisdiction_metrics[jur_id]; !ok {
		jurisdiction_metrics[jur_id] = &Metrics{jurisdiction: jur}
	}

	metrics := jurisdiction_metrics[jur_id]

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
		jurisdiction_metrics.applyAgentStateUpdate(parent, payload)
	}
}

func (jurisdiction_metrics JuristictionMetrics) reset() {
	for _, metrics := range jurisdiction_metrics {
		metrics.reset()
	}
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

	metrics.NewCases = 0
}

func (tx *MetricsTx) send(jurisdiction_metrics JuristictionMetrics) {
	routing_key := fmt.Sprintf("%s.%s", tx.api_id, tx.sim_id)

	body, err := json.Marshal(Notification{
		Type:    MetricsNotification,
		Payload: jurisdiction_metrics,
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
