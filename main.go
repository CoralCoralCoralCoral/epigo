package main

import (
	"log"
	"os"

	"github.com/CoralCoralCoralCoral/simulation-engine/messaging"
	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	rmq_conn, err := amqp091.Dial(os.Getenv("RMQ_URI"))
	if err != nil {
		log.Fatalf("couldn't create connection to rabbit: %s", err)
	}
	defer rmq_conn.Close()

	init_rx := messaging.NewInitRx(rmq_conn)
	init_rx.OnReceive(func(api_id uuid.UUID, config model.Config) {
		sim := model.NewSimulation(config)

		metrics_tx := messaging.NewMetricsTx(rmq_conn, api_id, sim.Id())
		defer metrics_tx.Close()

		sim.Subscribe(metrics_tx.NewEventSubscriber())

		command_rx := messaging.NewCommandRx(rmq_conn, sim.Id())
		defer command_rx.Close()

		go command_rx.OnReceive(sim.SendCommand)

		sim.Start()
	})
}
