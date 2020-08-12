package amqp

import (
	"os"
	"strconv"
)

type Config struct {
	Host  string `json:"host"`
	User  string `json:"user"`
	Pass  string `json:"pass"`
	VHost string `json:"v_host"`
	Port  int    `json:"port"`
}

const (
	EnvAmqpHost  = "AMQP_HOST"
	EnvAmqpUser  = "AMQP_USER"
	EnvAmqpPass  = "AMQP_PASS"
	EnvAmqpVHost = "AMQP_VHOST"
	EnvAmqpPort  = "AMQP_PORT"
)

func ParseWithEnv() *Config {

	conf := &Config{}

	for _, env := range []string{
		EnvAmqpHost,
		EnvAmqpUser,
		EnvAmqpPass,
		EnvAmqpVHost,
	} {
		val := os.Getenv(env)
		if val != "" {
			switch env {
			case EnvAmqpHost:
				conf.Host = val
				break
			case EnvAmqpPass:
				conf.Pass = val
				break
			case EnvAmqpUser:
				conf.User = val
				break
			case EnvAmqpVHost:
				conf.VHost = val
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
