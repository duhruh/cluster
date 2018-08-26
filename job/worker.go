package job

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Worker interface {
	Connect() error
	Popit(Messenger) error
	Close() error
}

type Messenger func(msg string) bool

type rabbitWorker struct {
	amqpURI    string
	connection *amqp.Connection
	queueName  string
}

func NewRabbitWorker(uri string) Worker {
	return &rabbitWorker{
		amqpURI:       uri,
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

func (r *rabbitWorker) Popit(m Messenger) error {
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
			if !m(string(d.Body)) {
				err := d.Nack(true, true)
				fmt.Printf("Nack: %v\n", err)
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
