package transit

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/transit"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-trans"
)

type DummyTransitter struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	transit func()
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name  string `validate:"required"`
	Group string `validate:"required"`
}

func init() {
	transit.Plugin[module] = &DummyTransitter{}
}

func (d *DummyTransitter) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.wg = &sync.WaitGroup{}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.transit = transit.WrapWithSingleMsgLoop(d.ctx, d.wg, d.Group, d.coreFunc)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyTransitter) CheckConfig() error {
	validator := validator.New()
	return validator.Struct(d.config)
}

func (d *DummyTransitter) coreFunc(byteTask []byte) (protocol.Job, error) {
	var task protocol.Job
	err := json.Unmarshal(byteTask, &task)
	if err != nil {
		d.logf.Error(err.Error())
		return protocol.Job{}, err
	}

	return task, nil
}

func (d *DummyTransitter) DoTransit() {
	d.transit()
}

func (d *DummyTransitter) Stop() {
	d.cancel()
	d.wg.Wait()
}
