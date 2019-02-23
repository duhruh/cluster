package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	Len  int    `json:"len"`
}

func myMessenger(message string) bool {
	var dp deliverPayload

	now := time.Now()

	err := json.Unmarshal([]byte(message), &dp)
	if err != nil {
		return false
	}

	hostname, err := os.Hostname()
	if err != nil {
		return false
	}
	fileName := fmt.Sprintf("/tmp/%s.txt", hostname)

	// download file
	out, err := os.Create(fileName)
	if err != nil {
		return false
	}

	resp, err := http.Get(dp.Line)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	out.Close()

	fh, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	nower := time.Now()
	for scanner.Scan() {
		substring, doesIt := doesStringWithMD5Exist(scanner.Text(), dp.MD5, dp.Len)
		if !doesIt {
			continue
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
	nowerr := time.Since(nower)
	println(fmt.Sprintf("[PARSE] took %v", nowerr))

	return true
}

func doesStringWithMD5Exist(line string, targetmd5 string, charLen int) (string, bool) {
	var strlen = len(line)
	for i := 0; i <= strlen; i++ {
		for j := i + 1; j <= strlen; j++ {
			ll := j - i
			if ll < charLen {
				continue
			}
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
