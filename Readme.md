# GO_AMQP

go的amqp简单封装库

- 断线重连
- 优雅重启
- 并发消费

## 怎么使用

````go
package main

import (
	"encoding/json"
	"github.com/iNightd/go-amqp"
	amqp_default_connect "github.com/iNightd/go-amqp/driver/amqp"
	"log"
	"sync"
	"testing"
	"time"
)

type Map map[string]interface{}

func main()  {
	amqp, err := go_amqp.New(
     	go_amqp.SetDriver(
            amqp_default_connect.New(&amqp_default_connect.Config{
                    User:  "root",
                    Pass:  "root",
                    VHost: "example",
                    Host:  "127.0.0.1",
                    Port:  5672,
            }),
        ),
    )
     
    printErr(err)

    queue, err := amqp.Queue("example", "example", "example")
    printErr(err)
    
	var wg sync.WaitGroup
    
    // 队列消费
    printErr(queue.Consume("reloadConsume", func(data []byte, name string) {
        log.Println(name, string(data[:]))
        wg.Done()
    }, 3))
    

    ticker := time.NewTicker(time.Second / 100)
    i := 0
    for range ticker.C {
        i++
        // 写入队列
        if err := queue.PublishJson(Map{
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

    wg.Wait()
}

func printErr(err error) {
    if err != nil {
        log.Println(err)
    }
}
````