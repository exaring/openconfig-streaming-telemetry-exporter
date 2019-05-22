package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metric struct {
	name   string
	labels []label
	value  interface{}
	desc   *prometheus.Desc
}

type label struct {
	key   string
	value string
}

func (m *metric) promName() string {
	return metricNameReplacer.Replace(m.name)
}

func (m *metric) describe() *prometheus.Desc {
	return prometheus.NewDesc(m.promName(), m.name, m.promLabelKeys(), nil)
}

func (m *metric) promLabelValues() []string {
	values := m.labelValues()
	res := make([]string, len(values))

	for i, v := range values {
		res[i] = labelValueReplacer.Replace(v)
	}

	return res
}

func (m *metric) labelValues() []string {
	res := make([]string, len(m.labels))
	for i, label := range m.labels {
		res[i] = label.value
	}

	return res
}

func (m *metric) promLabelKeys() []string {
	keys := m.labelKeys()
	res := make([]string, len(keys))

	for i, k := range keys {
		res[i] = labelKeyReplacer.Replace(k)
	}

	return res
}

func (m *metric) labelKeys() []string {
	res := make([]string, len(m.labels))
	for i, label := range m.labels {
		res[i] = label.key
	}

	return res
}
