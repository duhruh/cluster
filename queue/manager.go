package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Manager interface {
	Connect() error
	Publish([]byte) error
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

func NewRabbitMQManager(uri string) Manager {
	return &rabbitJobQueue{
		amqpURI:      uri,
		exchange:     "",
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

func (r *rabbitJobQueue) Publish(message []byte) error {
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
		r.exchange, // publish to an exchange
		q.Name,     // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            message,
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
	//	log.Printf("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		//		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}
