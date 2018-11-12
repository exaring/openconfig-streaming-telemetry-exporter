# openconfig-streaming-telemetry-exporter

openconfig-streaming-telemetry-exporter is a Prometheus exporter that collects metrics from network devices using streaming telemetry.

Tested with JunOS 17.3.

[![Build Status](https://travis-ci.org/exaring/openconfig-streaming-telemetry-exporter.svg?branch=master)](https://travis-ci.org/exaring/openconfig-streaming-telemetry-exporter)
[![Coverage Status](https://coveralls.io/repos/exaring/openconfig-streaming-telemetry-exporter/badge.svg?branch=master&service=github)](https://coveralls.io/github/exaring/openconfig-streaming-telemetry-exporter?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/exaring/openconfig-streaming-telemetry-exporter)](http://goreportcard.com/report/exaring/openconfig-streaming-telemetry-exporter)
[![Go Doc](https://godoc.org/github.com/exaring/openconfig-streaming-telemetry-exporter?status.svg)](https://godoc.org/github.com/exaring/openconfig-streaming-telemetry-exporter)

**PRs for other systems welcome!**

## Install

```go get github.com/exaring/openconfig-streaming-telemetry-exporter```

## Run

```openconfig-streaming-telemetry-exporter -config.file /path/to/config.yml```

## Configuration

```yaml
# The address to listen for Prometheus scrapers
listen_address: 0.0.0.0:9513
# Prometheus metrics path
metrics_path: /metrics
# Targets block, you can define multiple targets here
targets:
# Hostname of the openconfig target
- hostname: 203.0.113.1
  # Port of the openconfig target
  port: 50051
  # Openconfig paths to subscribe to
  paths:
    # Network interfaces metrics path
  - path: /interfaces/
    # Suppress updates for not changed metrics
    suppress_unchanged: false
    # Maximum time between updates
    max_silent_interval_ms: 20000
    # Sample frequency
    sample_frequency_ms: 2000
# As some metrics are returned as strings we need to map those to an int for Prometheus
string_value_mapping:
  # Path to do mappings for
  /interfaces/interface/state/oper-status:
    # string(DOWN) mapped to int(0)
    DOWN: 0
    # string(UP) mapped to int(1)
    UP: 1
```

## JunOS examples

### Device Configuration

<https://forums.juniper.net/t5/Automation/OpenConfig-and-gRPC-Junos-Telemetry-Interface/ta-p/316090>

### Available openconfig paths

<https://www.juniper.net/documentation/en_US/junos/topics/reference/general/junos-telemetry-interface-grpc-sensors.html>  
<https://www.juniper.net/documentation/en_US/junos/information-products/pathway-pages/open-config/open-config-feature-guide.pdf>

### Debug

If you want to see the internal tree structure you should visit /debug/dump.
General status information is available under /debug/state.
