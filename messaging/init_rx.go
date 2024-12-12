package messaging

import (
	"encoding/json"
	"log"

	"github.com/CoralCoralCoralCoral/simulation-engine/model"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type InitRx struct {
	ch       *amqp091.Channel
	messages <-chan amqp091.Delivery
}

func NewInitRx(conn *amqp091.Connection) *InitRx {
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")

	err = ch.ExchangeDeclare(INIT_EXCHANGE, "topic", false, true, false, false, nil)
	failOnError(err, "failed to create exchange")

	queue, err := ch.QueueDeclare(
		INIT_EXCHANGE, // Queue name
		false,         // Durable (survives broker restarts)
		false,         // Auto-delete
		false,         // Exclusive
		false,         // No-wait
		nil,           // Arguments
	)

	if err != nil {
		failOnError(err, "failed to declare init-game queue")
	}

	err = ch.QueueBind(queue.Name, "#", INIT_EXCHANGE, false, nil)
	failOnError(err, "failed to bind init-game queue to wildcard key on init-game exchange")

	// Set prefetch count to 1 for fair dispatch
	err = ch.Qos(1, 0, false)
	failOnError(err, "failed to set prefetch to 1")

	messages, err := ch.Consume(
		INIT_EXCHANGE, // Queue name
		"",            // Consumer tag
		false,         // Auto-acknowledge (set to false for manual acks)
		false,         // Exclusive
		false,         // No-local
		false,         // No-wait
		nil,           // Arguments
	)

	failOnError(err, "failed to consume from init channel")

	rx := InitRx{
		ch,
		messages,
	}

	return &rx
}

func (rx *InitRx) OnReceive(handler func(api_id uuid.UUID, config model.Config)) {
	defer rx.ch.Close()

	// debugging
	log.Println("listening for init messages")

	for msg := range rx.messages {
		// debugging
		log.Println("received init message")

		api_id, err := extractApiId(&msg)
		if err != nil {
			log.Println("couldn't extract api_id from init message")
			msg.Ack(false)
			continue
		}

		var config model.Config
		err = json.Unmarshal(msg.Body, &config)

		if err != nil {
			log.Println("failed to parse init message")
			msg.Ack(false)
			continue
		}

		go handler(api_id, config)

		msg.Ack(false)
	}
}
