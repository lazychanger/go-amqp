package aliyun_amqp

import (
	"os"
	"strconv"
)

type Config struct {
	Ak         string `json:"ak"`
	Sk         string `json:"sk"`
	InstanceId string `json:"instance_id"`
	Endpoint   string `json:"endpoint"`
	VHost      string `json:"v_host"`
	Port       int    `json:"port"`
}

const (
	EnvAmqpAk         = "AMQP_AK"
	EnvAmqpSk         = "AMQP_SK"
	EnvAmqpInstanceId = "AMQP_INSTANCE_ID"
	EnvAmqpEndpoint   = "AMQP_Endpoint"
	EnvAmqpVHost      = "AMQP_VHOST"
	EnvAmqpPort       = "AMQP_PORT"
)

func ParseWithEnv() *Config {

	conf := &Config{}

	for _, env := range []string{
		EnvAmqpAk,
		EnvAmqpEndpoint,
		EnvAmqpInstanceId,
		EnvAmqpSk,
		EnvAmqpVHost,
	} {
		val := os.Getenv(env)
		if val != "" {
			switch env {
			case EnvAmqpSk:
				conf.Sk = val
				break
			case EnvAmqpVHost:
				conf.Sk = val
				break
			case EnvAmqpEndpoint:
				conf.Endpoint = val
				break
			case EnvAmqpInstanceId:
				conf.InstanceId = val
				break
			case EnvAmqpAk:
				conf.Ak = val
				break
			case EnvAmqpPort:
				conf.Port, _ = strconv.Atoi(val)
				if conf.Port == 0 {
					conf.Port = 5672
				}
				break
			}
		}
	}
	return conf
}
