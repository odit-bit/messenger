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

	e := &rabbitClient{
		conn:         conn,
		exchangeName: _ExhchangeDefault,
	}
	ch, err := conn.Channel()
	panicOnError(err, "failed get channel")

	logMigrate(ch, e.exchangeName)
	ch.Close()
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

	return publisher{
		cli: e,
		ch:  ch,
	}
}

func (e *rabbitClient) Consumer(queue string) <-chan amqp.Delivery {
	var err error
	ch, err := e.conn.Channel()
	panicOnError(err, "failed set channel")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	panicOnError(err, "set Qos")

	msgC, err := ch.Consume(queue, "", false, false, false, false, nil)
	panicOnError(err, "failed consume channel")

	return msgC

}

func logMigrate(ch *amqp.Channel, exchange string) {
	migrate(ch, exchange,
		slog.LevelInfo.String(),
		slog.LevelError.String(),
		slog.LevelDebug.String(),
		slog.LevelWarn.String(),
	)
}

func migrate(ch *amqp.Channel, exchange string, routes ...string) {
	declareLogExchange(ch, exchange)

	//!!binding the exchange
	for _, key := range routes {
		q, err := ch.QueueDeclare(key, true, false, false, false, nil)
		panicOnError(err, "declare queue")

		err = ch.QueueBind(q.Name, key, exchange, false, nil)
		panicOnError(err, "failed bind")
		fmt.Printf("binding queue [%s] to exchange [%s], with key [%s] \n", q.Name, exchange, key)
	}

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
