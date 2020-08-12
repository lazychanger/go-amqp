package go_amqp

import "time"

type Option func(c *config)

func SetDriver(driver Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

func SetMaxSendRetries(maxSendRetries int) Option {
	return func(c *config) {
		c.maxSendRetries = maxSendRetries
	}
}

func SetMaxReconnect(maxReconnects int) Option {
	return func(c *config) {
		c.maxReconnects = maxReconnects
	}
}

func SetReconnectDelay(reconnectDelay time.Duration) Option {
	return func(c *config) {
		c.reconnectDelay = reconnectDelay
	}
}

func SetMaxOverstock(maxOverstock int64) Option {
	return func(c *config) {
		c.maxOverstock = maxOverstock
	}
}
