package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/duhruh/cluster/queue"
)

var (
	targetMD5 = flag.String("md5", "", "target md5 sum")
	charLen   = flag.Int("len", 10, "the char len")
)

func main() {
	flag.Parse()

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

const target = "69e5ac0cf03478d7e20b6fd3c7451c6f"
const fmtstring = "{\"line\":\"%s\",\"md5\":\"%s\",\"len\":%v}"

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

	length := len(files)
	for k, file := range files {

		println(fmt.Sprintf("you're at %s and %d/%d", file, k, length))
		uri := fmt.Sprintf("http://file-server%v", strings.Replace(file, "/outusb", "", 1))
		payload := fmt.Sprintf(fmtstring, uri, *targetMD5, *charLen)

		err = rabbit.Publish([]byte(payload))
		if err != nil {
			rabbit.Connect()
			println(err.Error())
		}

	}
	println("you have posted all the messages")
}
