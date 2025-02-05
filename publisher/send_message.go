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

func publishMessages(conn *amqp.Connection, queueName string, duration time.Duration, rate int, wg *sync.WaitGroup, generatedMessage string) {
	defer wg.Done()

	ch, err := conn.Channel() // Open a separate channel for each publisher
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	messageCount := 0
	byteMessage := []byte(generatedMessage)

	for {
		select {
		case <-ctx.Done():
			//log.Printf("Publisher finished. Total messages sent: %d", messageCount)
			return
		case <-ticker.C:
			for i := 0; i < rate; i++ {
				select {
				case <-ctx.Done():
					log.Printf("Publisher finished. Total messages sent: %d", messageCount)
					return
				default:
					//message := generateRandomMessage(messageSize)
					err := ch.PublishWithContext(ctx,
						"",        // exchange
						queueName, // routing key
						false,     // mandatory
						false,     // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        byteMessage,
							Headers: amqp.Table{
								"source":    "go_publisher",
								"timestamp": time.Now().Format(time.RFC3339Nano),
								"customKey": "customValue",
							},
						})
					if err != nil {
						log.Printf("Failed to publish message: %s", err)
					} else {
						messageCount++
					}
				}
			}
		}
	}
}

func main() {
	rabbitmqHost := flag.String("rabbitmqHost", "10.0.0.2", "Ip adress of the RabbitMQ Instance")
	queueName := flag.String("queueName", "transient_queue", "Name of the RabbitMQ queue")
	messageSize := flag.Int("messageSize", 1000, "Size of the messages in bytes")
	numPublishers := flag.Int("numPublishers", 10, "Number of concurrent publishers")
	testDuration := flag.Int("testDuration", 100, "Test duration in seconds")
	rate := flag.Int("rate", 100, "Rate of messages per second")
	// go run send_message.go -messageSize=1000 -numPublishers=100 -testDuration=20 -rate=1000 -queueName="transient_queue"

	message := generateRandomMessage(*messageSize)

	flag.Parse() // Parse the flags

	conn, err := amqp.Dial("amqp://daniel:daniel@" + *rabbitmqHost + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Start publishers
	var wg sync.WaitGroup
	for i := 0; i < *numPublishers; i++ {
		wg.Add(1)
		go publishMessages(conn, *queueName, time.Duration(*testDuration)*time.Second, *rate, &wg, message)
	}

	// Wait for all publishers to finish
	wg.Wait()
	log.Printf("Test wave completed for queue: %s\n", *queueName)
}
