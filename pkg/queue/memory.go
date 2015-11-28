package queue

import (
	"time"

	"github.com/eapache/channels"
	log "gopkg.in/inconshreveable/log15.v2"
)

type InfiniteRR struct {
	Index    int
	Channels []*channels.InfiniteChannel
}

type Memory struct {
	Queues map[string]map[string]*InfiniteRR
}

func NewMemory() (*Memory, error) {
	return &Memory{
		Queues: map[string]map[string]*InfiniteRR{},
	}, nil
}

type MemoryMessage struct {
	Message []byte
	Topic   string
	Channel string
	Queue   *Memory
}

func (m *MemoryMessage) GetBody() []byte {
	return m.Message
}

func (m *MemoryMessage) Requeue(delay time.Duration) {
	var err error
	if delay < 1 {
		err = m.Queue.Publish(m.Topic, m.Message)
	} else {
		err = m.Queue.DeferredPublish(m.Topic, delay, m.Message)
	}

	if err != nil {
		log.Error(
			"Unable to requeue a message",
			"topic", m.Topic,
			"channel", m.Channel,
			"delay", delay,
			"error", err,
		)
	}
}

func (m *MemoryMessage) Touch() {
	// noop
}

func (m *Memory) AddHandler(topic, channel string, handler Handler, concurrency int) error {
	if _, ok := m.Queues[topic]; !ok {
		m.Queues[topic] = map[string]*InfiniteRR{}
	}

	if _, ok := m.Queues[topic][channel]; !ok {
		m.Queues[topic][channel] = &InfiniteRR{
			Channels: []*channels.InfiniteChannel{},
		}
	}

	for i := 0; i < concurrency; i++ {
		ch := channels.NewInfiniteChannel()

		go func(ch *channels.InfiniteChannel) {
			for {
				select {
				case msg := <-ch.Out():
					if err := handler(&MemoryMessage{
						Message: msg.([]byte),
						Topic:   topic,
						Channel: channel,
						Queue:   m,
					}); err != nil {
						log.Error(
							"Error during queue handler execution",
							"topic", topic,
							"channel", channel,
							"error", err,
						)
					}
				}
			}
		}(ch)

		m.Queues[topic][channel].Channels = append(m.Queues[topic][channel].Channels, ch)
	}

	return nil
}

func (m *Memory) Publish(topic string, data []byte) error {
	if _, ok := m.Queues[topic]; !ok {
		return nil
	}

	for _, rr := range m.Queues[topic] {
		if rr == nil {
			continue
		}

		rr.Index++
		rr.Channels[rr.Index%len(rr.Channels)].In() <- data
	}

	return nil
}

func (m *Memory) DeferredPublish(topic string, delay time.Duration, data []byte) error {
	time.AfterFunc(delay, func() {
		if err := m.Publish(topic, data); err != nil {
			log.Error(
				"Unable to publish a deferred message",
				"topic", topic,
			)
		}
	})
	return nil
}

func (m *Memory) MultiPublish(topic string, data [][]byte) error {
	for _, body := range data {
		if err := m.Publish(topic, body); err != nil {
			return err
		}
	}

	return nil
}
