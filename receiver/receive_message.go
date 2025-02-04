package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	rabbitmqHost := flag.String("rabbitmqHost", "10.0.0.2", "Ip adress of the RabbitMQ Instance")
	queueName := flag.String("queueName", "transient_queue", "Name of the RabbitMQ queue")
	benchmarkID := "run_1" // Example benchmark ID

	flag.Parse() // Parse the flags

	conn, err := amqp.Dial("amqp://daniel:daniel@" + *rabbitmqHost + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	msgs, err := ch.Consume(
		*queueName,  // queue
		"consumer1", // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	failOnError(err, "Failed to register a consumer")

	// Open CSV file
	file, err := os.Create("benchmark_" + benchmarkID + ".csv")
	failOnError(err, "Failed to create CSV file")
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var forever chan struct{}

	var messageCount int

	go func() {
		for msg := range msgs {
			sentTime, ok := msg.Headers["timestamp"].(string)
			if !ok {
				log.Println("Timestamp header not found")
				continue
			}

			// Record current timestamp as received_time
			receivedTime := time.Now().Format(time.RFC3339Nano)

			// Increment message ID
			messageCount++

			// Write row to CSV
			err := writer.Write([]string{
				string(messageCount), // message_id
				sentTime,             // sent_time
				receivedTime,         // received_time
				benchmarkID,          // benchmark_id
				*queueName,           // queue_type
			})
			if err != nil {
				log.Printf("Failed to write to CSV: %s", err)
			}

			writer.Flush() // Ensure data is written to the file
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
