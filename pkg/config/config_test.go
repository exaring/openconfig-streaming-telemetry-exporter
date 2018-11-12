package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func boolAddr(v bool) *bool {
	return &v
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
