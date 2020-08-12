package aliyun_amqp

import (
	"bytes"
	"strconv"
)

type AliyunAmqp struct {
	Config *Config
}

func (a *AliyunAmqp) Url() string {
	var buf bytes.Buffer
	ak := a.Config.Ak
	sk := a.Config.Sk
	instanceId := a.Config.InstanceId // 请替换成您阿里云AMQP控制台首页instanceId

	userName := GetUserName(ak, instanceId)
	password := GetPassword(sk)
	buf.WriteString("amqp://")
	buf.WriteString(userName)
	buf.WriteString(":")
	buf.WriteString(password)

	// <Your End Point> 请从控制台获取。如果你使用的是杭州Region，那么Endpoint会形如 137000000010111.mq-amqp.cn-hangzhou-a.aliyuncs.com
	buf.WriteString("@")
	buf.WriteString(a.Config.Endpoint)
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

func New(config *Config) *AliyunAmqp {
	return &AliyunAmqp{
		Config: config,
	}
}
