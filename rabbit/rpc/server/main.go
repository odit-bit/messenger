package main

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func panicOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	panicOnError(err, "dial rabbit")
	defer conn.Close()

	ch, err := conn.Channel()
	panicOnError(err, "open channel")
	defer ch.Close()

	que, err := ch.QueueDeclare("job_queue", false, false, false, false, nil)
	panicOnError(err, "failed set queue")

	que2, err := ch.QueueDeclare("job_queue_2", false, false, false, false, nil)
	panicOnError(err, "failed set queue")

	err = ch.Qos(1, 0, false)
	panicOnError(err, "faile set Qos")

	msgC, err := ch.ConsumeWithContext(context.Background(), que.Name, "", false, false, false, false, nil)
	panicOnError(err, "failed register consumer")

	msg2C, err := ch.ConsumeWithContext(context.Background(), que2.Name, "", false, false, false, false, nil)
	panicOnError(err, "failed register consumer")

	//que
	var forever chan struct{}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		defer func() {
			forever <- struct{}{}
		}()

		for {
			var d amqp.Delivery
			var response string
			select {
			case d = <-msgC:
				fmt.Printf("got message on que: %v \n", string(d.Body))
				response = fmt.Sprintf("message receieved: %v, at: %v ", string(d.Body), d.Timestamp)

			case d = <-msg2C:
				fmt.Printf("got message on que2: %v \n", string(d.Body))
				response = fmt.Sprintf("message receieved: %v, at: %v ", string(d.Body), d.Timestamp)

			}

			err := ch.PublishWithContext(ctx, "", d.ReplyTo, false, false,
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				},
			)
			panicOnError(err, "failed to publish message")
			d.Ack(false)

		}

	}()

	fmt.Println("serving..")
	<-forever
}
