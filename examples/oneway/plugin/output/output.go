package output

import (
	"context"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/output"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-out"
)

type DummyOutputer struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	output func()
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name  string `validate:"required"`
	Group string `validate:"required"`
}

func init() {
	output.Plugin[module] = &DummyOutputer{}
}

func (d *DummyOutputer) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.wg = &sync.WaitGroup{}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.output = output.WrapWithSingleMsgLoop(d.ctx, d.wg, d.Group, d.coreFunc)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyOutputer) CheckConfig() error {
	validate := validator.New()
	return validate.Struct(d.config)
}

func (d *DummyOutputer) coreFunc(task protocol.Job) error {
	d.logf.Infof("dummy output demo: %s", task.String())
	return nil
}

func (d *DummyOutputer) DoOutput() {
	d.output()
}

func (d *DummyOutputer) Stop() {
	d.cancel()
	d.wg.Wait()
}
