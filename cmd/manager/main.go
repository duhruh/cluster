package main

import (
	"github.com/duhruh/cluster/job"
)

func main() {

	rabbit := job.NewRabbitWorker("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}

	forever := make(chan bool)
	err = rabbit.Popit(myMessenger)
	if err != nil {
		println(err.Error())
	}
	println("gonna wait for messages")
	<-forever

}

func myMessenger(message string) bool {
	println(message)

	return true
}