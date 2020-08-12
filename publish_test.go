package go_amqp

import (
	"encoding/json"
	"github.com/iNightd/go-amqp/driver/amqp"
	"log"
	"sync"
	"testing"
	"time"
)

type Map map[string]interface{}

func initQueue() (*Connection, *Queue, error) {
	a, err := New(
		SetDriver(amqp.New(&amqp.Config{
			User:  "root",
			Pass:  "root",
			VHost: "example",
			Host:  "127.0.0.1",
			Port:  5672,
		})),
	)

	if err != nil {
		log.Panic(a)
		return nil, nil, err
	}

	q, err := a.Queue("example", "example", "example")
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return a, q, nil
}

func TestPublish(t *testing.T) {
	a, q, _ := initQueue()
	var wg sync.WaitGroup
	var succeed = 0
	printErr(q.Consume("reloadConsume", func(data []byte, name string) {
		var body Map
		if err := json.Unmarshal(data, &body); err != nil {
			log.Println(name, err)
		} else {
			//log.Println(name, string(data[:]))
		}
		wg.Done()
		succeed++
	}, 3))
	// 控制速度
	ticker := time.NewTicker(time.Second / 100)
	i := 0
	for range ticker.C {
		i++
		if err := q.PublishJson(Map{
			"index": i,
			"now":   time.Now(),
		}); err != nil {
			log.Println(err)
		} else {
			wg.Add(1)
		}
		if i == 3000 {
			ticker.Stop()
		}
	}
	printErr(a.Close())
	wg.Wait()
	log.Println(succeed)
}

func printErr(err error) {
	if err != nil {
		log.Println(err)
	}
}