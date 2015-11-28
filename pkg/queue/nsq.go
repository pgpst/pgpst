package queue

import (
	"time"

	"github.com/nsqio/go-nsq"

	"code.pgp.st/pgpst/pkg/config"
)

type NSQ struct {
	LookupdAddresses []string
	NSQdAddresses    []string

	ProdRR    int
	Producers []*nsq.Producer
}

func NewNSQ(cfg config.NSQConfig) (*NSQ, error) {
	prs := []*nsq.Producer{}

	if cfg.ServerAddresses != nil {
		for _, addr := range cfg.ServerAddresses {
			pr, err := nsq.NewProducer(addr, nsq.NewConfig())
			if err != nil {
				return nil, err
			}
			if err := pr.Ping(); err != nil {
				return nil, err
			}

			prs = append(prs, pr)
		}
	}

	return &NSQ{
		LookupdAddresses: cfg.LookupdAddresses,
		NSQdAddresses:    cfg.ServerAddresses,
		Producers:        prs,
	}, nil
}

type NSQMessage struct {
	*nsq.Message
}

func (n *NSQMessage) GetBody() []byte {
	return n.Body
}

func (n *NSQ) AddHandler(topic, channel string, handler Handler, concurrency int) error {
	co, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	if err != nil {
		return err
	}

	co.AddConcurrentHandlers(nsq.HandlerFunc(func(msg *nsq.Message) error {
		return handler(&NSQMessage{msg})
	}), concurrency)

	if err := co.ConnectToNSQLookupds(n.LookupdAddresses); err != nil {
		return err
	}

	return nil
}

func (n *NSQ) Publish(topic string, data []byte) error {
	pr := n.Producers[n.ProdRR%len(n.Producers)]
	n.ProdRR++
	return pr.Publish(topic, data)
}

func (n *NSQ) DeferredPublish(topic string, delay time.Duration, data []byte) error {
	pr := n.Producers[n.ProdRR%len(n.Producers)]
	n.ProdRR++
	return pr.DeferredPublish(topic, delay, data)
}

func (n *NSQ) MultiPublish(topic string, data [][]byte) error {
	pr := n.Producers[n.ProdRR%len(n.Producers)]
	n.ProdRR++
	return pr.MultiPublish(topic, data)
}
