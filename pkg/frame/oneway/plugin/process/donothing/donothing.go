package donothing

import (
	"context"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/process"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
)

const (
	module = "donothing"
)

type Dazer struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	process func()
	Config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type Config struct {
	Name  string `validate:"required"`
	Group string `validate:"required"`
}

func init() {
	registerModule()
}

func registerModule() {
	process.Plugin[module] = &Dazer{
		wg: &sync.WaitGroup{},
	}
}

func (d *Dazer) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.Config)
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.process = process.WrapWithSingleMsgLoop(d.ctx, d.wg, d.Group, d.coreFunc)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *Dazer) CheckConfig() error {
	return nil
}

func (d *Dazer) coreFunc(job protocol.Job, isChanOpen bool) (protocol.Job, error) {
	return job, nil
}

func (d *Dazer) DoProcess() {
	d.process()
}

func (d *Dazer) Stop() {
	d.cancel()
	d.wg.Wait()
	d.logf.Infof("stop process plugin: %s", module)
}
