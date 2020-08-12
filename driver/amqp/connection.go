package amqp

import (
	"bytes"
	"strconv"
)

type Amqp struct {
	Config *Config
}

func (a *Amqp) Url() string {
	var buf bytes.Buffer

	buf.WriteString("amqp://")
	buf.WriteString(a.Config.User)
	buf.WriteString(":")
	buf.WriteString(a.Config.Pass)

	// <Your End Point> 请从控制台获取。如果你使用的是杭州Region，那么Endpoint会形如 137000000010111.mq-amqp.cn-hangzhou-a.aliyuncs.com
	buf.WriteString("@")
	buf.WriteString(a.Config.Host)
	buf.WriteString(":")
	if a.Config.Port == 0 {
		a.Config.Port = 5672
	}
	buf.WriteString(strconv.Itoa(a.Config.Port))
	buf.WriteString("/")
	buf.WriteString(a.Config.VHost)
	buf.WriteString("?timeout=10s")
	return buf.String()
}

func New(config *Config) *Amqp {
	return &Amqp{
		Config: config,
	}
}
