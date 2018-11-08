# prom-telemetry-gw

Prometheus Telemetry Gateways is a Prometheus exporter that collects metrics from network devices using streaming telemetry.

Currently tested with JunOS 17.3

## Example configuration on JunOS
https://forums.juniper.net/t5/Automation/OpenConfig-and-gRPC-Junos-Telemetry-Interface/ta-p/316090

## Install
```go get github.com/exaring/openconfig-streaming-telemetry-exporter```

## Run
```openconfig-streaming-telemetry-exporter -config.file /path/to/config.yml```

## Configuration
`listen_address: :8080` - The address to listen for Prometheus scrapers
`metrics_path: /metrics` - Path on which Prometheus metrics are exposed
`targets:` - Targets block, you can define multiple targets here
`- hostname: 169.254.0.0` - Hostname of the openconfig target
`  port: 50051` - Port of the openconfig target
`  paths:` - Openconfig paths to subscribe to
`  - path: /interfaces/` - Network interfaces metrics path
`    suppress_unchanged: false` - Surpress updates for not changed metrics
`    max_silent_interval_ms: 20000` - Maximum time between updates
`    sample_frequency_ms: 2000` - Sample frequency

As some metrics are returned as Strings we need to map those to an int for Prometheus. 
`string_value_mapping:` - Mapping of different String states to int values for Prometheus
`  /interfaces/interface/state/oper-status:` - Path to do mappings for
`    DOWN: 0` - string(DOWN) mapped to int(0)
`    UP: 1` - string(UP) mapped to int(1)
