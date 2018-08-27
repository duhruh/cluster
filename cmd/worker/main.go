package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/duhruh/cluster/queue"
)

func main() {

	rabbit := queue.NewRabbitMQWorker("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}

	forever := make(chan bool)
	err = rabbit.Popit(myMessenger)
	if err != nil {
		println(err.Error())
	}
	println("gonna wait for messages!")
	<-forever

}

type payload struct {
	Message string `json:"message"`
	Took    string `json:"took"`
}

type deliverPayload struct {
	Line string `json:"line"`
	MD5  string `json:"md5"`
}

func myMessenger(message string) bool {
	var dp deliverPayload

	err := json.Unmarshal([]byte(message), &dp)
	if err != nil {
		return false
	}

	now := time.Now()
	substring, doesIt := doesStringWithMD5Exist(dp.Line, dp.MD5)
	if !doesIt {
		return true
	}
	took := time.Since(now)

	println("http things")
	p := payload{
		Message: substring,
		Took:    fmt.Sprintf("%v", took),
	}
	bbs, err := json.Marshal(p)
	if err != nil {
		println(err.Error())
	}

	_, err = http.Post("http://manager:8080/report", "application/json", bytes.NewBuffer(bbs))
	if err != nil {
		println(err.Error())
	}

	return true
}

func doesStringWithMD5Exist(line string, targetmd5 string) (string, bool) {
	var strlen = len(line)
	for i := 0; i <= strlen; i++ {
		for j := i + 1; j <= strlen; j++ {
			substring := line[i:j]
			sum := md5.Sum([]byte(substring))
			sumstring := fmt.Sprintf("%x", sum)

			if sumstring == targetmd5 {
				return substring, true
			}

		}

	}

	return "", false
}
