package main

import (

	"time"
	"github.com/duhruh/cluster/queue"
)

func main() {

	rabbit := queue.NewRabbitMQManager("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}

	for true {
		err = rabbit.Publish("hello messages this could be something great")
		if err != nil {
			println(err.Error())
		}
		time.Sleep(5 * time.Second)
	}
	

}
