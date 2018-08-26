package main

import (
	"github.com/duhruh/cluster/job"
)

func main() {
	println("hello")

	rabbit := job.NewRabbitMQ("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}

	err = rabbit.Publish("hello messages this could be something great")
	if err != nil {
		println(err.Error())
	}

}
