package input

import (
	"context"
	"sync"
	"time"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/input"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-in"
)

type DummyInputter struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	input func()
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name          string `validate:"required"`
	Group         string `validate:"required"`
	FetchInterval int    `validate:"required"`
}

func init() {
	input.Plugin[module] = &DummyInputter{}
}

func (d *DummyInputter) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.wg = &sync.WaitGroup{}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.input = input.WrapWithSingleMsgLoop(d.ctx, d.wg, d.Group, d.coreFunc, time.Duration(d.FetchInterval))

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyInputter) CheckConfig() error {
	validate := validator.New()
	return validate.Struct(d.config)
}

func (d *DummyInputter) coreFunc() ([]byte, error) {
	return []byte(`{"task":"mock task"}`), nil
}

func (d *DummyInputter) DoInput() {
	d.input()
}

func (d *DummyInputter) Stop() {
	d.cancel()
	d.wg.Wait()
}
