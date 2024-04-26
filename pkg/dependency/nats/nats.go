package nats

import (
	"strconv"
	"strings"
	"time"

	planeLog "github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/nslookup"
	"github.com/bigstack-oss/plane-go/pkg/base/os"
	"github.com/nats-io/nats.go"
)

const (
	natsPrefix = "nats://"
)

var (
	log, logf = planeLog.GetLoggers("nats-helper")
)

type JetStreamClient interface {
	StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	AddStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error)
	Publish(string, []byte, ...nats.PubOpt) (*nats.PubAck, error)
	PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error)
}

type JetStreamSubscriber interface {
	Fetch(batch int, opts ...nats.PullOpt) ([]*nats.Msg, error)
}

type Helper struct {
	JsClient     JetStreamClient
	JsSubscriber JetStreamSubscriber

	Config
}

type Config struct {
	Sockets       []string `validate:"required"`
	IsHeadlessSvc bool
	Stream        string

	Subject  string
	Subjects map[string]string
	Durable  string

	Retention
	Fetch
}

type Retention struct {
	Second int
	Size   int
}

type Fetch struct {
	Interval time.Duration
	Number   int
	Retry    int
}

func (h *Helper) genHeadlessSockets(sockets []string) []string {
	socket := strings.ReplaceAll(sockets[0], natsPrefix, "")
	domainName := strings.Split(socket, ":")[0]
	port := strings.Split(socket, ":")[1]

	newSockets := []string{}
	ips := nslookup.ResolveIPs(domainName)
	for num := range ips {
		socket := natsPrefix + domainName + "-" + strconv.Itoa(num) + "." + domainName + ":" + port
		newSockets = append(newSockets, socket)
	}

	return newSockets
}

func (h *Helper) SetNatsJetStreamClient() {
	var err error
	scksStr := ""
	if h.IsHeadlessSvc {
		scksStr = strings.Join(h.genHeadlessSockets(h.Sockets), ",")
	} else {
		scksStr = strings.Join(h.Sockets, ",")
	}

	cli, err := nats.Connect(scksStr)
	if err != nil {
		logf.Errorf("error details of set nats connection: %s", err.Error())
		os.Exit(1)
	}

	h.JsClient, err = cli.JetStream()
	if err != nil {
		logf.Errorf("error details of set nats jetstream client: %s", err.Error())
		os.Exit(1)
	}
}

func (h *Helper) SetJetStreamSubscriber(subject string, durable string) {
	subscriber, err := h.JsClient.PullSubscribe(subject, durable)
	if err != nil {
		logf.Errorf("error details of set nats jetstream subscriber: %s", err.Error())
		return
	}

	h.JsSubscriber = subscriber
}

func (h *Helper) AddStream(stream string) error {
	_, err := h.JsClient.StreamInfo(stream)
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return err
	}

	streamPrefix := stream + ".*"
	_, err = h.JsClient.AddStream(
		&nats.StreamConfig{
			Name:      stream,
			Subjects:  []string{streamPrefix},
			Retention: nats.LimitsPolicy,
			MaxAge:    time.Duration(h.Retention.Second) * time.Second,
		},
	)

	return err
}

func (h *Helper) Publish(subject string, msg []byte) error {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return err
		}

		_, err = h.JsClient.Publish(subject, msg)
		if err != nil {
			trialCount++
			time.Sleep(2 * time.Second)
			continue
		}

		return nil
	}
}

func (h *Helper) PullSubscribe(subject string) ([][]byte, error) {
	var err error
	var trialCount int

	for {
		natsMsgs := []*nats.Msg{}
		if trialCount > h.Config.Retry {
			return nil, err
		}

		natsMsgs, err = h.JsSubscriber.Fetch(h.Fetch.Number)
		if err != nil {
			trialCount++
			continue
		}

		var msgs [][]byte
		for _, natsMsg := range natsMsgs {
			msgs = append(msgs, natsMsg.Data)
			natsMsg.Ack()
		}

		return msgs, nil
	}
}
