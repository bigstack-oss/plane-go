package metric

import (
	"encoding/json"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/monitoring"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	module    = "metric"
	oneMinute = 60 * time.Second
)

var (
	inputOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "input_ok",
			Help: "",
		},
	)

	inputErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "input_err",
			Help: "",
		},
	)

	transitOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transit_ok",
			Help: "",
		},
	)

	transitErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transit_err",
			Help: "",
		},
	)

	processOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_ok",
			Help: "",
		},
	)

	processErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_err",
			Help: "",
		},
	)

	outputOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "output_ok",
			Help: "",
		},
	)

	outputErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "output_err",
			Help: "",
		},
	)

	metricLogger = log.GetLogger(module)
)

func RegisterGaugeMetric(gauge *prometheus.GaugeVec) {
	monitoring.MetricRegistry.MustRegister(gauge)
}

func showMetrics(metrics *plugin.Metric) {
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		metricLogger.Error("failed to dump metrics periodically")
	} else {
		rawMessage := json.RawMessage(jsonMetrics)
		metricLogger.Info("periodical metrics dump", zap.Any("metrics", &rawMessage))
	}
}

func swapMetrics(metrics **plugin.Metric) *plugin.Metric {
	oldMetrics := &plugin.Metric{}
	newMetrics := &plugin.Metric{}

	*(*unsafe.Pointer)(unsafe.Pointer(&oldMetrics)) = atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(metrics)),
		*(*unsafe.Pointer)(unsafe.Pointer(&newMetrics)),
	)

	return oldMetrics
}

func setMetrics(metrics *plugin.Metric) {
	inputOK.Set(float64(metrics.InputOK))
	inputErr.Set(float64(metrics.InputErr))

	transitOK.Set(float64(metrics.TransitOK))
	transitErr.Set(float64(metrics.TransitErr))

	processOK.Set(float64(metrics.ProcessOK))
	processErr.Set(float64(metrics.ProcessErr))

	outputOK.Set(float64(metrics.OutputOK))
	outputErr.Set(float64(metrics.OutputErr))
}
