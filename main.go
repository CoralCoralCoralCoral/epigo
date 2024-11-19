package main

import (
	"github.com/CoralCoralCoralCoral/simulation-engine/messaging"
	"github.com/CoralCoralCoralCoral/simulation-engine/metrics"
	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func main() {

	// create a game with 150k people
	sim := model.NewSimulation(model.Config{
		Id:        uuid.New(),
		NumAgents: 150000,
		TimeStep:  15 * 60 * 1000,
		Pathogen: model.Pathogen{
			IncubationPeriod:   [2]float64{3 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
			RecoveryPeriod:     [2]float64{7 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
			ImmunityPeriod:     [2]float64{330 * 24 * 60 * 60 * 1000, 90 * 24 * 60 * 60 * 1000},
			QuantaEmissionRate: [2]float64{250, 100},
		},
	})

	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic("couldn't connect to rabbit")
	}

	// start a new metrics instance subscribed to simulation events
	sim.Subscribe(metrics.NewEventSubscriber(conn, uuid.Max, sim.Id()))
	sim.Subscribe(messaging.NewGameUpdateTx(conn, uuid.Max, sim.Id()).NewEventSubscriber())

	sim.Start()
}
