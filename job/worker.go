package job

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Worker interface {
	Connect() error
	Popit() (string, error)
	Close() error
}

type Messenger func(msg string) bool

type rabbitWorker struct {
	amqpURI    string
	connection *ampq.Connection
	queueName  string
}

func NewRabbitWorker(uri string) Worker {
	return &rabbitWorker{
		uri:       uri,
		queueName: "swarm-queue",
	}
}

func (r *rabbitWorker) Connect() error {
	connection, err := amqp.Dial(r.ampqURI)
	if err != nil {
		return fmt.Errorf("Dial: %v", err)
	}
	r.connection = connection

	return nil
}

func (r *rabbitWorker) Popit(m Messenger) (string, error) {
	ch, err := r.connection.Channel()
	if err != nil {
		return "", nil
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
		return "", fmt.Errorf("QueueDeclare: %v", err)
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
		return "", fmt.Errorf("Consume: %v", err)
	}

	go func() {
		for d := range msg {
			if !m(d.Body) {
				err := d.Nack(true, true)
				fmt.Printf("Nack: %v\n", err)
				continue
			}
			d.Ack(false)
		}
	}()
}

func (r *rabbitWorker) Close() error {
	return r.connection.Close()
}
