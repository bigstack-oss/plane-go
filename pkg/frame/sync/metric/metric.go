package metric

import (
	"encoding/json"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/monitoring"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	module    = "metric"
	oneMinute = 60 * time.Second
)

var (
	interactOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "interact_ok",
			Help: "",
		},
	)

	interactErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "interact_err",
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

	requestOK = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "request_ok",
			Help: "",
		},
	)

	requestErr = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "request_err",
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
		return
	}

	rawMessage := json.RawMessage(jsonMetrics)
	metricLogger.Info("periodical metrics dump", zap.Any("metrics", &rawMessage))
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
	interactOK.Set(float64(plugin.Metrics.InteractOK))
	interactErr.Set(float64(plugin.Metrics.InteractErr))

	transitOK.Set(float64(plugin.Metrics.TransitOK))
	transitErr.Set(float64(plugin.Metrics.TransitErr))

	processOK.Set(float64(plugin.Metrics.ProcessOK))
	processErr.Set(float64(plugin.Metrics.ProcessErr))

	requestOK.Set(float64(plugin.Metrics.RequestOK))
	requestErr.Set(float64(plugin.Metrics.RequestErr))
}
