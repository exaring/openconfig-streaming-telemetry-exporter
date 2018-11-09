package collector

import (
	"fmt"
	"strings"
	"sync"

	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector is a streaming telemetry data collector
type Collector struct {
	cfg       *config.Config
	targets   map[string]*Target
	targetsMu sync.RWMutex
}

// New initializes a new Collector
func New(cfg *config.Config) *Collector {
	return &Collector{
		cfg:     cfg,
		targets: make(map[string]*Target),
	}
}

// Stop stops the collector
func (c *Collector) Stop() {
	c.targetsMu.RLock()
	defer c.targetsMu.RUnlock()

	for _, t := range c.targets {
		t.stop()
	}
}

func (c *Collector) Dump() []string {
	c.targetsMu.RLock()
	defer c.targetsMu.RUnlock()

	for _, t := range c.targets {
		return t.dump()
	}

	return nil
}

// AddTarget adds a target to the collector
func (c *Collector) AddTarget(tconf *config.Target, stringValueMapping map[string]map[string]int) *Target {
	c.targetsMu.Lock()
	defer c.targetsMu.Unlock()
	c.targets[tconf.Hostname] = newTarget(tconf, stringValueMapping)

	return c.targets[tconf.Hostname]
}

// Describe is required by prometheus interface
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
}

// Collect collects data from the collector and send it to prometheus
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.targetsMu.RLock()
	defer c.targetsMu.RUnlock()

	var wg sync.WaitGroup
	for _, t := range c.targets {
		wg.Add(1)
		go t.collect(ch, &wg)
	}

	wg.Wait()
}

func parseKey(input string) (string, []label, error) {
	data := []rune(input)

	key := make([]rune, 0, 40)
	labelStrings := make([]string, 0, 3)
	withinAngledBraces := false
	tmp := make([]rune, 0, 10)

	for i := 0; i < len(data); i++ {
		if !withinAngledBraces {
			if data[i] == '[' {
				withinAngledBraces = true
				continue
			}

			key = append(key, data[i])
			continue
		}

		if data[i] == ']' {
			labelStrings = append(labelStrings, string(tmp))
			withinAngledBraces = false
			tmp = make([]rune, 0)
			continue
		}

		tmp = append(tmp, data[i])
	}

	retLabels := make([]label, len(labelStrings))
	for i, labelStr := range labelStrings {
		l, err := parseLabel(labelStr)
		if err != nil {
			return "", nil, fmt.Errorf("Unable to parse label: %v", err)
		}

		retLabels[i] = *l
	}

	return string(key), retLabels, nil

}

func parseLabel(input string) (*label, error) {
	parts := strings.Split(input, "=")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Unexpected length of split string: %q: %d", input, len(parts))
	}

	return &label{
		key:   parts[0],
		value: strings.Replace(parts[1], "'", "", -1),
	}, nil
}
