package go_amqp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lazychanger/go-amqp/tools"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type Queue struct {
	sync.Mutex
	// 队列名
	name string
	// 交换机名
	exchangeName string
	// 队列绑定交换机路由KEY
	routingKey string
	// 交换机模式
	exchange string
	// 最大内置队列长度
	maxOverstock int64

	conn    *amqp.Connection
	channel *amqp.Channel

	// 1.为什么不采用slice。因为golang的slice我们不方便作为队列操作，每次入出都会引发大量的内存操作。所以采用 map 类型
	// 2.为什么采用sync.map。因为入与出是同时执行，异步操作，如果直接采用map类型，会导致脏读、幻读等问题，我们需要对map类型的读写需要锁，而golang有内置完整的读写锁map模型
	// 3.采用map类型以后，需要我们自己维护头出尾入以及长度，所以这里有qMaxI、qMinI、ql进行数据记录
	q *sync.Map
	// 最大index
	qMaxI int64
	// 最小index
	qMinI int64
	// 内置队列长度
	ql int64
	// 是否启动释放库存
	isReleaseStock bool
	// 是否重启消费
	isReloadConsume int
	// 是否停止消费
	isStopConsume bool
	// 是否准备好
	already int

	close bool
	// 消费者
	cs []consume
}

type MessageStatus int8

const (
	AlreadyStop    = 0
	AlreadyReload  = 1
	AlreadySucceed = 2

	MessageStatusSucceed MessageStatus = 1
	MessageStatusError   MessageStatus = 2
	MessageStatusRequeue MessageStatus = 3
)

// 提供重启服务
func (q *Queue) Reload(conn *amqp.Connection) error {
	// 先关闭状态
	q.conn = conn
	// 标识重启中
	q.already = AlreadyReload
	return q.init()
}

// 初始化操作
func (q *Queue) init() error {
	var (
		ch  *amqp.Channel
		err error
	)
	log.Println("[amqp] queue init start")
	q.Lock()
	ch, err = q.conn.Channel()
	if err != nil {
		return err
	}

	// 创建交换机
	if err = ch.ExchangeDeclare(q.exchangeName, q.exchange, false, false, false, false, nil); err != nil {
		return errors.New(fmt.Sprintf("[amqp] exchange declare failed, err: %s", err))
	}
	// 创建队列
	if _, err = ch.QueueDeclare(q.name, false, false, false, false, nil); err != nil {
		return errors.New(fmt.Sprintf("[amqp] queue declare failed, err: %s", err))
	}
	// 交换机绑定队列
	if err = ch.QueueBind(q.name, q.routingKey, q.exchangeName, false, nil); err != nil {
		return errors.New(fmt.Sprintf("[amqp] queue bind failed, err: %s", err))
	}
	// 管道交换
	q.channel = ch
	// 告知已经准备完毕
	if q.already == AlreadyReload {
		q.already = AlreadySucceed

		// 重新触发消费
		q.reloadConsume()

		if !q.isReleaseStock && q.ql > 0 {
			log.Println("init release stock")
			// 重新触发
			go q.releaseStock()
		}
	}

	q.Unlock()

	log.Println("[amqp] queue init end")

	return nil
}

// 对推送简单封装一下，使json对象推送更加简便
func (q *Queue) PublishJson(v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return q.Publish(body)
}

// 原始推送，要求[]byte
func (q *Queue) Publish(data []byte) error {
	// 检查库存，防止溢出
	if q.maxOverstock > 0 && q.ql > q.maxOverstock {
		return publishOverstock
	}

	if q.close {
		return errors.New("service closing")
	}

	// 启动锁，防止释放库存时候，共同操作导致脏写
	q.Lock()
	// 防止并发map创建
	if q.q == nil {
		q.q = &sync.Map{}
	}
	// 增加最大下标
	q.qMaxI++
	// 增加最大长度
	q.ql++
	// 存储值
	q.q.Store(q.qMaxI, data)
	q.Unlock()

	// 检查释放库存是否启动，未启动并且channel已经准备完毕，就启动库存释放
	if !q.isReleaseStock && q.already == AlreadySucceed {
		go q.releaseStock()
	}

	log.Printf("published，now %d", q.ql)
	return nil
}

// 内置队列消费，库存释放
func (q *Queue) releaseStock() {
	// 判断是否重复启动
	q.Lock()
	if q.isReleaseStock {
		q.Unlock()
		log.Println("[amqp] release stock already run")
		return
	}
	// 标记服务已经启动
	q.isReleaseStock = true
	q.Unlock()
	log.Println("[amqp] release stock")
	for {
		// 如果库存为空或者channel还未准备好就关闭循环
		if q.ql == 0 || q.already != AlreadySucceed {
			break
		}
		// 先将当前长度取出，防止循环时候修改，变成脏读
		l := q.ql
		// 实际启动当前轮次库存释放
		for i := int64(0); i < l; i++ {
			if q.already != AlreadySucceed {
				break
			}
			// 对库存
			q.Lock()
			log.Printf("internal queues length: %d", q.ql)
			// 对库存最小下标进行+1
			q.qMinI++
			// 减少库存最大数
			q.ql--
			// 锁期间，顶部索引不变
			min := q.qMinI
			q.Unlock()

			// 读取内容
			body, has := q.q.Load(min)
			// 预防脏读
			if has && body != nil {
				// 推送
				_ = q.publish(body.([]byte))
			} else {
				log.Println("[amqp] data error")
			}
			// 释放map空间
			q.q.Delete(min)
		}

		// 本轮库存释放已经结束，延迟执行3秒后执行下一轮
		ticker := time.NewTicker(time.Second * 3)
		select {
		case <-ticker.C:
			ticker.Stop()
		}
	}
	// 标记关闭
	q.isReleaseStock = false
}

// 实际channel向队列发送数据
func (q *Queue) publish(data []byte) error {
	return q.channel.Publish(q.exchangeName, q.routingKey, false, false, amqp.Publishing{
		Body: data,
	})
}

// 添加消费内容，先存储，等待服务启动以后触发
func (q *Queue) Consume(name string, consumeFunc ConsumeFunc, repeat int) error {
	q.cs = append(q.cs, consume{
		repeat:      tools.IF(repeat <= 0, 1, repeat).(int),
		consumeFunc: consumeFunc,
		name:        name,
	})
	// 尝试启动消费
	q.reloadConsume()

	return nil
}

// 暂停消费
func (q *Queue) StopConsume() {
	q.isStopConsume = true
}

// 启动消费
func (q *Queue) StartConsume() {
	q.isStopConsume = false
	q.reloadConsume()
}

// 实际触发消费
func (q *Queue) reloadConsume() {
	// 如果未启动，直接返回
	if q.already != AlreadySucceed || q.isStopConsume {
		return
	}
	// 推送重启消费
	q.isReloadConsume++
	// 记录当前消费重启值
	reloadConsume := q.isReloadConsume
	// 记录当前channel重启值
	for i, c := range q.cs {
		// 并发消费
		for l := 0; l < c.repeat; l++ {
			name := fmt.Sprintf("%s_%d-%d", c.name, i, l)
			msgs, err := q.channel.Consume(q.name, name, false, false, false, false, nil)
			if err != nil {
				log.Fatalf("[AMQP] customer register err;name: %s, %s", name, err)
			} else {
				go func(c ConsumeFunc, consumeName string, reloadConsume int) {
					for msg := range msgs {
						switch c(msg.Body, consumeName) {
						case MessageStatusSucceed:
						case MessageStatusError:
							_ = msg.Ack(true)
							break
						case MessageStatusRequeue:
							_ = msg.Reject(true)
							break
						}
						// 如果channel重启或者消费重启，都结束当前消费，防止溢出，或者正在关闭
						if q.already != AlreadySucceed || q.isReloadConsume != reloadConsume || q.close || q.isStopConsume {
							break
						}
					}
				}(c.consumeFunc, name, reloadConsume)
			}
		}
	}
}

// 优雅重启
func (q *Queue) Close() error {
	// 先标记关闭
	q.close = true

	retry := 0

	for {

		if q.ql > 0 {
			if q.already == AlreadySucceed {
				if q.isReleaseStock == false {
					q.releaseStock()
				}
			} else {
				retry++
				// 如果channel没有准备好，内置队列也没有释放完，则重试三次，三次还没有处理好，就放弃重试
				if retry > 3 {
					break
				}
			}
		} else {
			break
		}

		ticker := time.NewTicker(time.Second / 2)
		select {
		case <-ticker.C:
			ticker.Stop()
		}
	}

	return q.channel.Close()
}

type consume struct {
	name        string
	consumeFunc ConsumeFunc
	repeat      int
}
type ConsumeFunc func(data []byte, name string) MessageStatus
