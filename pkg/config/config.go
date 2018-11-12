package config

import (
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultListenAddress       = ":9513"
	defaultMetricsPath         = "/metrics"
	defaultKeepaliveSeconds    = 1
	defaultTimeoutFactor       = 3
	defaultSampleFrequencyMS   = 5000
	defaultMaxSilentIntervalMS = 15000
	defaultSuppressUnchanged   = true
)

// Config is the configuration of the prom-telemetry-gw
type Config struct {
	ListenAddress      string                    `yaml:"listen_address"`
	MetricsPath        string                    `yaml:"metrics_path"`
	Targets            []*Target                 `yaml:"targets"`
	StringValueMapping map[string]map[string]int `yaml:"string_value_mapping"`
	Version            string
}

// Target represents a monitored system
type Target struct {
	Hostname   string  `yaml:"hostname"`
	Port       uint16  `yaml:"port"`
	KeepaliveS uint16  `yaml:"keepalive_s"`
	TimeoutS   uint16  `yaml:"timeout_s"`
	Paths      []*Path `yaml:"paths"`
}

// Path represents a resource identifier, e.g. /junos/system/linecard/cpu/memory/
// See https://www.juniper.net/documentation/en_US/junos/topics/reference/configuration-statement/sensor-edit-services-analytics.html
// for more examples.
type Path struct {
	Path                string `yaml:"path"`
	SuppressUnchanged   *bool  `yaml:"suppress_unchanged"`
	MaxSilentIntervalMS uint64 `yaml:"max_silent_interval_ms"`
	SampleFrequencyMS   uint64 `yaml:"sample_frequency_ms"`
}

// New creates a new empty config object
func New() *Config {
	return &Config{}
}

// Load loads a config from reader
func Load(reader io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	c := New()
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	c.LoadDefaults()
	return c, nil
}

// LoadDefaults loads default settings for config c
func (c *Config) LoadDefaults() {
	if c.ListenAddress == "" {
		c.ListenAddress = defaultListenAddress
	}

	if c.MetricsPath == "" {
		c.MetricsPath = defaultMetricsPath
	}

	for i := range c.Targets {
		if c.Targets[i].KeepaliveS == 0 {
			c.Targets[i].KeepaliveS = defaultKeepaliveSeconds
		}

		if c.Targets[i].TimeoutS == 0 {
			c.Targets[i].TimeoutS = defaultTimeoutFactor * c.Targets[i].KeepaliveS
		}

		for j := range c.Targets[i].Paths {
			if c.Targets[i].Paths[j].SampleFrequencyMS == 0 {
				c.Targets[i].Paths[j].SampleFrequencyMS = defaultSampleFrequencyMS
			}

			if c.Targets[i].Paths[j].MaxSilentIntervalMS == 0 {
				c.Targets[i].Paths[j].MaxSilentIntervalMS = defaultMaxSilentIntervalMS
			}

			if c.Targets[i].Paths[j].SuppressUnchanged == nil {
				x := defaultSuppressUnchanged
				c.Targets[i].Paths[j].SuppressUnchanged = &x
			}
		}
	}
}
