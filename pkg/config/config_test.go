package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func boolAddr(v bool) *bool {
	return &v
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Config
	}{
		{
			name: "Test #1",
			input: `
listen_address: 0.0.0.0:9513
metrics_path: /metrics
targets:
  - hostname: 203.0.113.1
    port: 50051
    keepalive_s: 1
    timeout_s: 3
    paths:
    - path: /interfaces/
      suppress_unchanged: false
      max_silent_interval_ms: 20000
      sample_frequency_ms: 2000
`,
			expected: &Config{
				ListenAddress: "0.0.0.0:9513",
				MetricsPath:   "/metrics",
				Targets: []*Target{
					{
						Hostname:   "203.0.113.1",
						Port:       50051,
						KeepaliveS: 1,
						TimeoutS:   3,
						Paths: []*Path{
							{
								Path:                "/interfaces/",
								SuppressUnchanged:   boolAddr(false),
								SampleFrequencyMS:   2000,
								MaxSilentIntervalMS: 20000,
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		cfg, _ := Load(bytes.NewReader([]byte(test.input)))
		assert.Equal(t, test.expected, cfg, test.name)
	}
}

func TestLoadDefaults(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected *Config
	}{
		{
			name: "Test #1",
			cfg: &Config{
				ListenAddress: "",
				MetricsPath:   "",
				Targets: []*Target{
					{
						KeepaliveS: 0,
						TimeoutS:   0,
						Paths: []*Path{
							{
								SampleFrequencyMS:   0,
								MaxSilentIntervalMS: 0,
								SuppressUnchanged:   nil,
							},
						},
					},
				},
			},
			expected: &Config{
				ListenAddress: defaultListenAddress,
				MetricsPath:   defaultMetricsPath,
				Targets: []*Target{
					{
						KeepaliveS: 1,
						TimeoutS:   3,
						Paths: []*Path{
							{
								SampleFrequencyMS:   defaultSampleFrequencyMS,
								MaxSilentIntervalMS: defaultMaxSilentIntervalMS,
								SuppressUnchanged:   boolAddr(defaultSuppressUnchanged),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.cfg.LoadDefaults()
		assert.Equal(t, test.expected, test.cfg, test.name)
	}
}
