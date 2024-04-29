package cronjob

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/cronjob"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "dummy-cron"
)

type DummyCronner struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	schedule func()
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name     string `validate:"required"`
	Schedule string `validate:"required"`
}

func init() {
	cronjob.Plugin[module] = &DummyCronner{
		wg: &sync.WaitGroup{},
	}
}

func (d *DummyCronner) SetConfig(conf map[string]interface{}) {
	_ = mapstructure.Decode(conf, &d.config)

	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.schedule = cronjob.WrapWithCron(d.ctx, d.wg, d.Schedule, d.coreFunc)

	d.log = log.GetLogger(module)
	d.logf = d.log.Sugar()
}

func (d *DummyCronner) CheckConfig() error {
	validate := validator.New()
	return validate.Struct(d.config)
}

func (d *DummyCronner) swapRecords(metrics **plugin.Metric) *plugin.Metric {
	oldMetrics := &plugin.Metric{}
	newMetrics := &plugin.Metric{}

	*(*unsafe.Pointer)(unsafe.Pointer(&oldMetrics)) = atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(metrics)),
		*(*unsafe.Pointer)(unsafe.Pointer(&newMetrics)),
	)

	return oldMetrics
}

func (d *DummyCronner) coreFunc() {
	metrics := d.swapRecords(&plugin.Metrics)
	jsonMetrics, _ := json.Marshal(metrics)
	raw := json.RawMessage(jsonMetrics)

	d.log.Info("dummy periodical metrics dump", zap.Any("metrics", &raw))
}

func (d *DummyCronner) DoSchedule() {
	d.schedule()
}

func (d *DummyCronner) Stop() {
	d.cancel()
	d.wg.Wait()
	d.ctx, d.cancel = context.WithCancel(context.Background())
}
