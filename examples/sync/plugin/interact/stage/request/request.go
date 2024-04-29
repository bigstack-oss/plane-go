package request

import (
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/stage"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-request"
)

type DummyRequester struct {
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name string
}

func init() {
	stage.Plugins[module] = &DummyRequester{}
}

func (d *DummyRequester) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyRequester) CheckConfig() error {
	validate := validator.New()
	return validate.Struct(d.config)
}

func (d *DummyRequester) Execute(task *protocol.Job) (bool, error) {
	d.logf.Info(d.Name)
	return true, nil
}
