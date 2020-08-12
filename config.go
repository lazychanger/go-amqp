package go_amqp

import "time"

type config struct {
	// 连接AMQP DSN构建驱动
	driver Driver

	// 最大消息重发次数
	maxSendRetries int
	// 最大重连次数
	maxReconnects int
	// 重连延迟时间
	reconnectDelay time.Duration

	// 最大发送积压数
	maxOverstock int64
}
