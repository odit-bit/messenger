package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/ksuid"
)

func panicOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func jobRPC(msg string, que string) (string, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	panicOnError(err, "failed dial rabbit")
	defer conn.Close()

	ch, err := conn.Channel()
	panicOnError(err, "Failed to open a channel")
	defer ch.Close()

	msgC, err := ch.Consume(
		que,   // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	panicOnError(err, "Failed to register a consumer")
	corrId := ksuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx, "", que, false, false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       que,
			Body:          []byte(msg),
		},
	)
	panicOnError(err, "failed publish message")

	var res string
	for d := range msgC {
		// if corrID is not same it is not our request
		if corrId == d.CorrelationId {
			res = string(d.Body)
			panicOnError(err, "Failed to convert body to integer")
			break
		}
	}

	return res, err
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for {
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		cmd := strings.Split(text, "@")
		if len(cmd) != 2 {
			panicOnError(fmt.Errorf("need only 2 argument"), "send job message")
		}
		rep, err := jobRPC(cmd[0], cmd[1])
		panicOnError(err, "jobRPC")
		fmt.Printf(">> %v \n", rep)

	}
}
