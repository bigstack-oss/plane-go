package process

import (
	"context"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/metric"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/process"
	"github.com/goinggo/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-proc"
)

var (
	customMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_custom_metric",
			Help: "",
		},
		[]string{
			"scenario_id",
		},
	)
)

type DummyProcessor struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	process func()
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name  string `validate:"required"`
	Group string `validate:"required"`
	ID    string
}

func init() {
	process.Plugin[module] = &DummyProcessor{}
	metric.RegisterGaugeMetric(customMetric)
}

func (d *DummyProcessor) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.wg = &sync.WaitGroup{}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.process = process.WrapWithSingleMsgLoop(d.ctx, d.wg, d.Group, d.coreFunc)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyProcessor) CheckConfig() error {
	validate := validator.New()
	return validate.Struct(d.config)
}

func (d *DummyProcessor) coreFunc(task protocol.Job, isChnOpen bool) (protocol.Job, error) {
	d.logf.Infof("dummy process demo: %s", task.String())
	customMetric.WithLabelValues("fakescenairoIDooxx").Set(6.66)

	return task, nil
}

func (d *DummyProcessor) DoProcess() {
	d.process()
}

func (d *DummyProcessor) Stop() {
	d.cancel()
	d.wg.Wait()
}
