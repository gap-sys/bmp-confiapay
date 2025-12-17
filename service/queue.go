package service

import "time"

type Queue interface {
	Produce(queue string, data any, delay time.Duration) error
}
