package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type CommandRx struct {
	ch       *amqp091.Channel
	messages <-chan amqp091.Delivery
}

func NewCommandRx(conn *amqp091.Connection, sim_id uuid.UUID) *CommandRx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare(COMMAND_EXCHANGE, "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	queue_name := sim_id.String()
	routing_key := fmt.Sprintf("*.%s", sim_id.String())

	_, err = ch.QueueDeclare(
		queue_name, // Queue name
		false,      // Durable (survives broker restarts)
		false,      // Auto-delete
		false,      // Exclusive
		false,      // No-wait
		nil,        // Arguments
	)

	if err != nil {
		log.Fatalf("Failed to declare command queue for simulation %s: %s", sim_id.String(), err)
	}

	err = ch.QueueBind(
		queue_name,       // Queue name
		routing_key,      // Routing key (matches all messages intended for the simulation with id = sim_id)
		COMMAND_EXCHANGE, // Exchange name
		false,            // No-wait
		nil,              // Arguments
	)

	if err != nil {
		log.Fatalf("Failed to bind command queue to routing key: %s", err)
	}

	messages, err := ch.Consume(
		queue_name, // Queue name
		"",         // Consumer tag
		false,      // Auto-acknowledge (set to false for manual acks)
		false,      // Exclusive
		false,      // No-local
		false,      // No-wait
		nil,        // Arguments
	)

	failOnError(err, "failed to consume from command channel")

	rx := CommandRx{
		ch,
		messages,
	}

	return &rx
}

func (rx *CommandRx) OnReceive(handler func(command model.Command)) {
	for msg := range rx.messages {
		var command model.Command
		err := json.Unmarshal(msg.Body, &command)

		if err != nil {
			log.Println("failed to parse command message")
			continue
		}

		handler(command)

		msg.Ack(false) // Acknowledge message
	}
}

func (rx *CommandRx) Close() {
	rx.ch.Close()
}
