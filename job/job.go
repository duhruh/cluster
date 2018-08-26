package job

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type JobQueue interface {
	Connect() error
	Publish(string) error
	Close() error
}

type rabbitJobQueue struct {
	amqpURI      string
	connection   *amqp.Connection
	exchange     string
	exchangeType string
	routingKey   string
	queueName    string
}

func NewRabbitMQ(uri string) JobQueue {
	return &rabbitJobQueue{
		amqpURI: uri,
		//		exchange:     "swarm-exchange",
		exchangeType: "direct",
		routingKey:   "swarm-work",
		queueName:    "swarm-queue",
	}
}

func (r *rabbitJobQueue) Connect() error {
	connection, err := amqp.Dial(r.amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	r.connection = connection

	return nil
}

func (r *rabbitJobQueue) Publish(message string) error {
	channel, err := r.connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	q, err := channel.QueueDeclare(
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

	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	defer confirmOne(confirms)

	if err = channel.Publish(
		r.exchange,   // publish to an exchange
		r.routingKey, // routing to 0 or more queues
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(message),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}
	return nil
}

func (r *rabbitJobQueue) Close() error {
	return r.connection.Close()
}

func confirmOne(confirms <-chan amqp.Confirmation) {
	log.Printf("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}
