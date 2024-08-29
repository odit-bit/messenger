package rlog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ///
const (
	_ExhchangeDefault = "log-default"
	_QueueDefault     = "log_queue"
)

var (
	WarnLevel  = slog.LevelWarn.String()
	InfoLevel  = slog.LevelInfo.String()
	DebugLevel = slog.LevelDebug.String()
)

// declare exchange on rabbitMQ server
func declareLogExchange(ch *amqp.Channel, exchangeName string) {
	err := ch.ExchangeDeclare(
		exchangeName,        // name
		amqp.ExchangeDirect, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	panicOnError(err, "rabbitMQ failed to declare exchange")
}

// wrap the amqp conn configured for this package use
type rabbitClient struct {
	conn         *amqp.Connection
	exchangeName string
}

func NewRabbit(uri string) *rabbitClient {

	conn, err := amqp.Dial(uri)
	panicOnError(err, "failed to dial server")

	ch, err := conn.Channel()
	panicOnError(err, "failed set channel")

	e := &rabbitClient{
		conn:         conn,
		exchangeName: _ExhchangeDefault,
	}

	declareLogExchange(ch, e.exchangeName)
	return e

}

func (e *rabbitClient) Close() error {
	var err error
	err = errors.Join(err, e.conn.Close())
	return err
}

func (e *rabbitClient) Publisher() publisher {
	ch, err := e.conn.Channel()
	panicOnError(err, "emitter failed set channel")

	declareLogExchange(ch, e.exchangeName)

	// the method's channel will close this
	err = ch.Confirm(false)
	panicOnError(err, "emitter failed set confirmation")

	notifC := make(chan amqp.Return, 1)
	confirmation := ch.NotifyReturn(notifC)

	go func() {
		for r := range confirmation {
			fmt.Printf("log publish, time: %v\n", r.Timestamp)
		}
	}()

	return publisher{
		cli: e,
		ch:  ch,
	}
}

func (e *rabbitClient) Consume(keyRoutes ...string) <-chan []byte {
	ch, err := e.conn.Channel()
	panicOnError(err, "failed set channel")

	declareLogExchange(ch, e.exchangeName)

	q, err := ch.QueueDeclare(_QueueDefault, false, false, false, false, nil)
	panicOnError(err, "failed declare")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	panicOnError(err, "")

	//!!binding the exchange
	for _, key := range keyRoutes {
		err = ch.QueueBind(q.Name, key, e.exchangeName, false, nil)
		panicOnError(err, "failed bind")
		fmt.Printf("binding to exchange %s, with key %s \n", e.exchangeName, key)
	}

	msgC, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	panicOnError(err, "failed consume channel")

	inC := make(chan []byte)
	go func() {
		defer close(inC)
		for d := range msgC {
			inC <- d.Body
			d.Ack(false)

		}
	}()

	return inC
}

// Publisher

type publisher struct {
	cli *rabbitClient
	ch  *amqp.Channel
}

func (pub *publisher) Close() error {
	return pub.ch.Close()
}

func (pub *publisher) Publish(ctx context.Context, body []byte, key string) error {
	err := pub.ch.PublishWithContext(ctx, pub.cli.exchangeName, key, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		return err
	}
	return nil
}

// // consumer

// type consumer struct {
// 	cli  *rabbitClient
// 	msgC <-chan []byte
// }

// func (c *consumer) Message(ctx context.Context) ([]byte, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	case msg := <-c.msgC:
// 		return msg, nil
// 	}
// }
