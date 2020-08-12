package go_amqp

import "errors"

var (
	missingConnectionDsnDriver = errors.New("missing connection dsn driver")
	alreadyClosed              = errors.New("already closed")

	publishOverstock  = errors.New("queue publish overstock")
	consumeNameRepeat = errors.New("reloadConsume name repeat")
)
