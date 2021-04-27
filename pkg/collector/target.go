package collector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/config"
	pb "github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	backoffInit = time.Second
	backoffMax  = time.Second * 16
)

// Target represents a streaming telemetry exporting network device
type Target struct {
	address            string
	devName            string
	con                *grpc.ClientConn
	client             pb.OpenConfigTelemetryClient
	paths              []*config.Path
	metrics            *tree
	stringValueMapping map[string]map[string]int
	stopCh             chan struct{}
	reconnect          bool
	maxReads           int
}

func newTarget(tconf *config.Target, stringValueMapping map[string]map[string]int, reconnect bool) *Target {
	t := &Target{
		address:            fmt.Sprintf("%s:%d", tconf.Hostname, tconf.Port),
		devName:            tconf.Hostname,
		paths:              tconf.Paths,
		metrics:            newTree(tconf.Hostname),
		stringValueMapping: stringValueMapping,
		reconnect:          reconnect,
	}

	return t
}

func (t *Target) stop() {
	t.stopCh <- struct{}{}
}

func (t *Target) dump() []string {
	return t.metrics.dump()
}

func (t *Target) subscriptionRequest() *pb.SubscriptionRequest {
	subReq := &pb.SubscriptionRequest{
		AdditionalConfig: &pb.SubscriptionAdditionalConfig{
			LimitRecords:     -1,
			LimitTimeSeconds: -1,
			NeedEos:          false,
		},
		PathList: make([]*pb.Path, 0),
	}

	for _, p := range t.paths {
		subReq.PathList = append(subReq.PathList, &pb.Path{
			Path:              p.Path,
			SuppressUnchanged: *p.SuppressUnchanged,
			MaxSilentInterval: uint32(p.MaxSilentIntervalMS),
			SampleFrequency:   uint32(p.SampleFrequencyMS),
		})
	}

	return subReq
}

func (t *Target) subscribe(con *grpc.ClientConn) pb.OpenConfigTelemetry_TelemetrySubscribeClient {
	backoff := time.Duration(0)

	for {
		select {
		case <-t.stopCh:
			return nil
		default:
		}

		time.Sleep(backoff)
		cl := pb.NewOpenConfigTelemetryClient(con)
		stream, err := cl.TelemetrySubscribe(context.Background(), t.subscriptionRequest())
		if err != nil {
			log.Errorf("TelemetrySubscribe failed: %v", err)

			if backoff == 0 {
				backoff = backoffInit
				continue
			}

			if backoff < backoffMax {
				backoff = backoff * 2
				continue
			}

			continue
		}

		return stream
	}
}

func (t *Target) process(stream pb.OpenConfigTelemetry_TelemetrySubscribeClient) {
	i := 0

	for {
		select {
		case <-t.stopCh:
			return
		default:
		}

		data, err := stream.Recv()
		if err != nil {
			log.Errorf("Failed to receive stream from [%v]: %v", t.devName, err)
			t.metrics = newTree(t.devName)
			break
		}

		t.processOpenConfigData(data)

		i++
		if t.maxReads > 0 && i >= t.maxReads {
			return
		}
	}
}

// Serve is the main handling routine for a target
func (t *Target) Serve(con *grpc.ClientConn) {
	defer con.Close()

	for {
		stream := t.subscribe(con)
		if stream == nil {
			return
		}

		t.process(stream)

		if !t.reconnect {
			return
		}
	}
}

func (t *Target) processOpenConfigData(data *pb.OpenConfigData) {
	prefix := ""
	for _, kv := range data.Kv {
		if kv.Key == "__prefix__" {
			if kv.Value == nil {
				log.Warningf("Received __prefix__ key with nil value from %s", t.address)
				prefix = ""
				continue
			}

			switch value := kv.Value.(type) {
			case *pb.KeyValue_StrValue:
				prefix = value.StrValue
			default:
				log.Warningf("Received __prefix__ key with non string value from %s", t.address)
			}

			continue
		}

		if strings.HasPrefix(kv.Key, "__") {
			continue
		}

		if strings.HasSuffix(kv.Key, "state/description") {
			if kv.Value == nil {
				continue
			}

			switch value := kv.Value.(type) {
			case *pb.KeyValue_StrValue:
				t.metrics.setDescription(strings.Replace(prefix+kv.Key, "state/description", "", -1), value.StrValue)
			}
		}

		t.metrics.insert(prefix+kv.Key, kv.Value)
	}
}

func (t *Target) collect(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	res := t.metrics.getMetrics()
	for _, m := range res {
		if m.value == nil {
			continue
		}

		valueType := prometheus.GaugeValue
		if strings.Contains(m.name, "counters") {
			valueType = prometheus.CounterValue
		}

		v := float64(0)
		switch value := m.value.(type) {
		case *pb.KeyValue_DoubleValue:
			v = float64(value.DoubleValue)
		case *pb.KeyValue_IntValue:
			v = float64(value.IntValue)
		case *pb.KeyValue_UintValue:
			v = float64(value.UintValue)
		case *pb.KeyValue_SintValue:
			v = float64(value.SintValue)
		case *pb.KeyValue_BoolValue:
			if value.BoolValue {
				v = 1
			}
		case *pb.KeyValue_StrValue:
			if _, ok := t.stringValueMapping["/"+m.name]; !ok {
				continue
			}

			if _, ok := t.stringValueMapping["/"+m.name][value.StrValue]; !ok {
				continue
			}

			v = float64(t.stringValueMapping["/"+m.name][value.StrValue])
		default:
			log.Fatalf("Unknown data type for %v", value)
		}

		ch <- prometheus.MustNewConstMetric(m.desc, valueType, v, m.promLabelValues()...)
	}
}
