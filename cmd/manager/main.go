package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/duhruh/cluster/queue"
)

func main() {

	var wg sync.WaitGroup

	wg.Add(2)

	go processUSBShit(wg)

	go httpReporter(wg)

	wg.Wait()
}

func httpReporter(wg sync.WaitGroup) {
	defer wg.Done()

	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			println(err.Error())
		}
		defer r.Body.Close()

		println(string(body))
	})
	println("http server listening on port :8080")
	http.ListenAndServe(":8080", nil)
}

const target = "b6fe428350544bea67513196957573d1"
const fmtstring = "{\"line\":\"%s\",\"md5\":\"%s\"}"

func processUSBShit(wg sync.WaitGroup) {
	defer wg.Done()
	rabbit := queue.NewRabbitMQManager("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}

	file, err := os.Open("/outusb/nnnnn.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = rabbit.Publish([]byte(fmt.Sprintf(fmtstring, scanner.Text(), target)))
		if err != nil {
			println(err.Error())
		}
	}
	println("if you haven't found it by now you never will")
}
