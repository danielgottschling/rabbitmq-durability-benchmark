package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
	}
}

func main() {
	// Get the RabbitMQ connection string from an environment variable
	// host := os.Getenv("RABBITMQ_HOST")
	// if host == "" {
	// 	log.Fatal("RABBITMQ_HOST environment variable is not set")
	// }

	host := "34.40.14.203"

	conn, err := amqp.Dial("amqp://daniel:daniel@" + host + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	msgs, err := ch.Consume(
		"transient_queue", // queue
		"consumer1",       // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			body := string(msg.Body)
			log.Printf("Received message: %s", body)

			// Access headers
			if timestamp, ok := msg.Headers["timestamp"]; ok {
				log.Printf("Message Timestamp: %v", timestamp)
			} else {
				log.Println("Timestamp header not found")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
