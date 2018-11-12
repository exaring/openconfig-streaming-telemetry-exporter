package collector

type metricSet struct {
	metrics []metric
}

func newMetricSet() *metricSet {
	return &metricSet{
		metrics: make([]metric, 0, 1000),
	}
}

func (ms *metricSet) append(m metric) {
	ms.metrics = append(ms.metrics, m)
}

func (ms *metricSet) get() []metric {
	return ms.metrics
}
