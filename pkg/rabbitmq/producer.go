package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

var conn *amqp091.Connection
var channel *amqp091.Channel

func Connect() error {
	var err error
	conn, err = amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}

	_, err = channel.QueueDeclare(
		"image_processing", // Queue name
		true,               // Durable
		false,              // Auto delete
		false,              // Exclusive
		false,              // No-wait
		nil,                // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	return nil
}

func PublishMessage(message string) error {
	if channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	err := channel.Publish(
		"",                 // Exchange
		"image_processing", // Routing key
		false,              // Mandatory
		false,              // Immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}
