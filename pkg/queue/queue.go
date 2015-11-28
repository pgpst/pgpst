package queue

import (
	"time"
)

type Queue interface {
	AddHandler(string, string, Handler, int) error
	Publish(string, []byte) error
	DeferredPublish(string, time.Duration, []byte) error
	MultiPublish(string, [][]byte) error
}

type Handler func(Message) error
type Message interface {
	GetBody() []byte
	Requeue(time.Duration)
	Touch()
}
