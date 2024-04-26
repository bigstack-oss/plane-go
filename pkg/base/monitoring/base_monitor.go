package monitoring

type Metricer interface {
	Report()
}

type Monitor struct {
	Metricer
}

func getMonitor(isOneTimeExec bool) Metricer {
	if isOneTimeExec {
		return GetMetricPusher()
	}

	return GetMetricPuller()
}

func (m *Monitor) SetReportTunnel(isOneTimeExec bool) {
	m.Metricer = getMonitor(isOneTimeExec)
}

func (m *Monitor) TraceMetric() {
	go m.Metricer.Report()
}
