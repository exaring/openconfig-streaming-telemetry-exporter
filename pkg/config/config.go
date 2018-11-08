package config

import (
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
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
	Hostname string  `yaml:"hostname"`
	Port     uint16  `yaml:"port"`
	Paths    []*Path `yaml:"paths"`
}

// Path represents a resource identifier, e.g. /junos/system/linecard/cpu/memory/
// See https://www.juniper.net/documentation/en_US/junos/topics/reference/configuration-statement/sensor-edit-services-analytics.html
// for more examples.
type Path struct {
	Path                string `yaml:"path"`
	SuppressUnchanged   bool   `yaml:"suppress_unchanged"`
	MaxSilentIntervalMS uint16 `yaml:"max_silent_interval_ms"`
	SampleFrequencyMS   uint16 `yaml:"sample_frequency_ms"`
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

	return c, nil
}
