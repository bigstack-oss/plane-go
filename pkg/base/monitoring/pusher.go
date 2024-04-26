package monitoring

import (
	"time"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	module    = "metricsPusher"
	pushURL   = "http://127.0.0.1:9091"
	oneMinute = 60 * time.Second
)

var (
	pusherLogger   = log.GetLogger(module).Sugar()
	MetricRegistry = prometheus.NewRegistry()
)

type Pusher interface {
	Add() error
	Gatherer(prometheus.Gatherer) *push.Pusher
}

type MetricPusher struct {
	Pusher
}

func GetMetricPusher() Metricer {
	return &MetricPusher{
		Pusher: push.New(pushURL, plugin.Service).Gatherer(MetricRegistry),
	}
}

func (m *MetricPusher) Report() {
	for {
		err := m.Pusher.Add()
		if err != nil {
			pusherLogger.Errorf("failed to push metric to metric proxy: %s", err.Error())
		}

		time.Sleep(oneMinute)
	}
}
