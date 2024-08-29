package rlog

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type emitter struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	exchangeName string
}

func NewEmitter(uri string) *emitter {
	conn, err := amqp.Dial(uri)
	panicOnError(err, "emitter failed to dial server")

	ch, err := conn.Channel()
	panicOnError(err, "emitter failed set channel")

	e := &emitter{
		conn:         conn,
		ch:           ch,
		exchangeName: _ExhchangeDefault,
	}

	declareLogExchange(ch, e.exchangeName)
	return e

}

func (e *emitter) Close() error {
	var err error
	err = errors.Join(err, e.ch.Close())
	err = errors.Join(err, e.conn.Close())
	return err
}

func (e *emitter) Publish(body []byte, key string) error {
	err := e.ch.Publish(e.exchangeName, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		return err
	}
	return nil
}
