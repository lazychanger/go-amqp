package go_amqp

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type Connection struct {
	// 配置
	config *config
	// amqp连接
	conn *amqp.Connection
	// 连接状态
	isConnected bool
	// 关机提示
	done chan bool
	// amqp闪断通知
	notifyClose chan *amqp.Error
	// 多队列（channel），此处认为一个channel管理一条队列
	qs []*Queue
}

func New(opts ...Option) (*Connection, error) {
	// 建立默认连接对象
	conn := &Connection{
		// 生成默认配置
		config: &config{
			reconnectDelay: time.Second,
			maxReconnects:  10,
			maxSendRetries: 3,
		},
	}

	for _, opt := range opts {
		// 配置写入
		opt(conn.config)
	}

	// 基础必须配置检查
	if conn.config.driver == nil {
		return nil, missingConnectionDsnDriver
	}

	// 断线重连
	go conn.handleReconnect()

	return conn, nil
}

func (c *Connection) handleReconnect() {
	tryConnects := 0
	for {
		c.isConnected = false
		if err := c.connect(); err != nil {
			if c.config.maxReconnects > 0 && tryConnects > c.config.maxReconnects {
				log.Fatalf("[AMQP] Reconnection times exceeded！(%d)", tryConnects)
				return
			}

			tryConnects += 1
			log.Printf("Failed to connect. %s Retrying...(%d)", err, tryConnects)
			time.Sleep(c.config.reconnectDelay)
			continue
		} else {
			// clear try connect
			tryConnects = 0
		}
		select {
		case <-c.done:
			return
		case <-c.notifyClose:
		}
	}
}

//
func (c *Connection) connect() error {
	url := c.config.driver.Url()
	log.Printf("[amqp] connected. %s", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	c.conn = conn

	c.notifyClose = make(chan *amqp.Error)
	c.conn.NotifyClose(c.notifyClose)

	c.isConnected = true
	c.queueReconnect()
	return nil
}

// channel重连
// 当连接闪断以后，需要重新建立新的连接，所以，所有的channel也需要进行新的连接
func (c *Connection) queueReconnect() {
	for _, q := range c.qs {
		if err := q.Reload(c.conn); err != nil {
			log.Println(err)
		}
	}
}

// 优雅关闭，
// 关闭channel
// 关闭闪断重连机制
// 关闭connection
func (c *Connection) Close() error {
	if c.isConnected {
		return alreadyClosed
	}
	var wg sync.WaitGroup
	// 批量关闭
	for i := 0; i < len(c.qs); i++ {
		wg.Add(1)
		go func(idx int, group *sync.WaitGroup) {
			_ = c.qs[idx].Close()
			group.Done()
		}(i, &wg)
	}

	wg.Wait()

	_ = c.conn.Close()

	// 关闭
	close(c.done)
	return nil
}

// 生成单Channel管理，让一个channel管理一条队列的发送与消费
func (c *Connection) Queue(name, exchange, routingKey string) (*Queue, error) {
	q := &Queue{
		name:         name,
		exchange:     amqp.ExchangeDirect,
		exchangeName: exchange,
		routingKey:   routingKey,
		maxOverstock: c.config.maxOverstock,
	}

	if c.isConnected == true {
		if err := q.Reload(c.conn); err != nil {
			return nil, err
		}
	}
	c.qs = append(c.qs, q)
	return q, nil
}
