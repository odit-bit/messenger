package rlog

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type receiver struct {
	conn         *amqp.Connection
	keyRoute     []string
	exchangeName string
	isBind       bool
}

func NewReceiver(uri string, route ...string) *receiver {
	conn, err := amqp.Dial(uri)
	panicOnError(err, "failed dial rabbit")

	return &receiver{
		conn:         conn,
		keyRoute:     route,
		exchangeName: _ExhchangeDefault,
	}
}

func (r *receiver) AddLevelSeverity(level string) {
	if r.isBind {
		panicOnError(fmt.Errorf("receiver binded"), "failed add severity")
	} else {
		r.keyRoute = append(r.keyRoute, level)
	}
}

func (r *receiver) Close() error {
	var err error
	err = errors.Join(err, r.conn.Close())
	return err
}

func (r *receiver) Receive() <-chan []byte {
	ch, err := r.conn.Channel()
	panicOnError(err, "failed set channel")

	declareLogExchange(ch, r.exchangeName)

	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	panicOnError(err, "failed declare")

	//!!binding the exchange
	for _, key := range r.keyRoute {
		err = ch.QueueBind(q.Name, key, r.exchangeName, false, nil)
		panicOnError(err, "failed bind")
		fmt.Printf("binding to exchange %s, with key %s \n", r.exchangeName, key)
	}
	r.isBind = true

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
