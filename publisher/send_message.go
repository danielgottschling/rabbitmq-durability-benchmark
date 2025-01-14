package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func generateRandomMessage(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	message := make([]byte, length)
	for i := range message {
		message[i] = charset[rand.Intn(len(charset))]
	}
	return string(message)
}

func publishMessages(ch *amqp.Channel, queueName string, messageSize int, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			message := generateRandomMessage(messageSize)
			err := ch.PublishWithContext(ctx,
				"",        // exchange
				queueName, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(message),
					Headers: amqp.Table{
						"source":    "go_publisher",
						"timestamp": time.Now().Format(time.RFC3339Nano),
						"customKey": "customValue",
					},
				})
			if err != nil {
				log.Printf("Failed to publish message: %s", err)
			}
		}
	}
}

func main() {
	rabbitmqHost := flag.String("rabbitmqHost", "10.0.0.2", "Ip adress of the RabbitMQ Instance")
	queueName := flag.String("queueName", "transient_queue", "Name of the RabbitMQ queue")
	messageSize := flag.Int("messageSize", 1, "Size of the messages in bytes")
	numPublishers := flag.Int("numPublishers", 10, "Number of concurrent publishers")
	testDuration := flag.Int("testDuration", 30, "Test duration in seconds")

	flag.Parse() // Parse the flags

	conn, err := amqp.Dial("amqp://daniel:daniel@" + *rabbitmqHost + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Start publishers
	var wg sync.WaitGroup
	for i := 0; i < *numPublishers; i++ {
		wg.Add(1)
		go publishMessages(ch, *queueName, *messageSize, time.Duration(*testDuration)*time.Second, &wg)
	}

	// Wait for all publishers to finish
	wg.Wait()
	log.Printf("Test wave completed for queue: %s\n", *queueName)
}
