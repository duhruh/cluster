package queue

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Worker interface {
	Connect() error
	Popit(Job) error
	Close() error
}

type Job func(msg string) bool

type rabbitWorker struct {
	amqpURI    string
	connection *amqp.Connection
	queueName  string
}

func NewRabbitMQWorker(uri string) Worker {
	return &rabbitWorker{
		amqpURI:   uri,
		queueName: "swarm-queue",
	}
}

func (r *rabbitWorker) Connect() error {
	connection, err := amqp.Dial(r.amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %v", err)
	}
	r.connection = connection

	return nil
}

func (r *rabbitWorker) Popit(m Job) error {
	ch, err := r.connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("QueueDeclare: %v", err)
	}

	msg, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("Consume: %v", err)
	}

	go func() {
		for d := range msg {
			if d.Redelivered {
				d.Reject(false)
				continue
			}
			if !m(string(d.Body)) {
				d.Nack(true, true)
				//fmt.Printf("Nack: %v\n", err)
				continue
			}
			d.Ack(false)
		}
	}()

	return nil
}

func (r *rabbitWorker) Close() error {
	return r.connection.Close()
}
