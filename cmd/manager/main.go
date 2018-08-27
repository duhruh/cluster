package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

const target = "59bb5beee293c5ee42638b8d73dc6676"
const fmtstring = "{\"line\":\"%s\",\"md5\":\"%s\"}"

func processUSBShit(wg sync.WaitGroup) {
	defer wg.Done()
	rabbit := queue.NewRabbitMQManager("amqp://guest:guest@rabbit-manage:5672/")

	err := rabbit.Connect()
	if err != nil {
		println(err.Error())
	}
	///media/pi/72B7-554A/round1/
	directory := "/outusb/round1"

	files := []string{}
	filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".txt") {
			return nil
		}
		files = append(files, path)
		return nil
	})

	for _, file := range files {
		fileHandle, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer fileHandle.Close()
		scanner := bufio.NewScanner(fileHandle)
		i := 0
		for scanner.Scan() {
			if i%2000 == 0 {
				time.Sleep(1 * time.Second)
			}
			err = rabbit.Publish([]byte(fmt.Sprintf(fmtstring, scanner.Text(), target)))
			if err != nil {
				rabbit.Connect()
				println(err.Error())
			}
			i++
		}

	}
	println("if you haven't found it by now you never will")
}
