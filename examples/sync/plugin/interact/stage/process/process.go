package process

import (
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/stage"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-process"
)

type DummyProcessor struct {
	validator *validator.Validate

	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name string
}

func init() {
	stage.Plugins[module] = &DummyProcessor{}
}

func (d *DummyProcessor) SetConfig(conf interface{}) {
	d.validator = validator.New()

	_ = mapstructure.Decode(conf, &d.config)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyProcessor) CheckConfig() error {
	return d.validator.Struct(d.config)
}

func (d *DummyProcessor) Execute(task *protocol.Job) (bool, error) {
	d.logf.Info(d.Name)
	return true, nil
}
