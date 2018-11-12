package collector

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/config"
	pb "github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/telemetry"
)

type mockTelemetryServer struct {
	testdata []pb.OpenConfigData
	wg       sync.WaitGroup
	stop     chan struct{}
}

func (m *mockTelemetryServer) TelemetrySubscribe(req *pb.SubscriptionRequest, srv pb.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	for _, test := range m.testdata {
		srv.Send(&test)
	}

	m.wg.Done()
	<-m.stop
	return nil
}

func (m *mockTelemetryServer) CancelTelemetrySubscription(context.Context, *pb.CancelSubscriptionRequest) (*pb.CancelSubscriptionReply, error) {
	return nil, nil
}

func (m *mockTelemetryServer) GetTelemetrySubscriptions(context.Context, *pb.GetSubscriptionsRequest) (*pb.GetSubscriptionsReply, error) {
	return nil, nil
}

func (m *mockTelemetryServer) GetTelemetryOperationalState(context.Context, *pb.GetOperationalStateRequest) (*pb.GetOperationalStateReply, error) {
	return nil, nil
}

func (m *mockTelemetryServer) GetDataEncodings(context.Context, *pb.DataEncodingRequest) (*pb.DataEncodingReply, error) {
	return nil, nil
}

func TestIntegration(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		testdata []pb.OpenConfigData
		expected string
	}{
		{
			name: "Test #1",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "some_label=somevalue,another-label=foobar",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{another_label=\"foobar\",name=\"xe-0/0/0\",some_label=\"somevalue\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{another_label=\"foobar\",name=\"xe-0/0/0\",some_label=\"somevalue\"} 1337\n",
		},
		{
			name: "Test #2",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "some totally random description",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{name=\"xe-0/0/0\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{name=\"xe-0/0/0\"} 1337\n",
		},
		{
			name: "Test #3",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "bar=foo",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
						{
							Key: "interface[name='xe-0/0/1']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "foo=bar",
							},
						},
						{
							Key: "interface[name='xe-0/0/1']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/1']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{foo=\"bar\",name=\"xe-0/0/1\"} 100\ninterfaces_interface_state_oper_state{bar=\"foo\",name=\"xe-0/0/0\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{foo=\"bar\",name=\"xe-0/0/1\"} 1337\ninterfaces_interface_state_pkts{bar=\"foo\",name=\"xe-0/0/0\"} 1337\n",
		},
		{
			name: "Test #4",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "bar=foo",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/subinterfaces/subinterface[index='123']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "some=label",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/subinterfaces/subinterface[index='123']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 232323,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{bar=\"foo\",name=\"xe-0/0/0\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{bar=\"foo\",name=\"xe-0/0/0\"} 1337\n# HELP interfaces_interface_subinterfaces_subinterface_state_pkts interfaces/interface/subinterfaces/subinterface/state/pkts\n# TYPE interfaces_interface_subinterfaces_subinterface_state_pkts gauge\ninterfaces_interface_subinterfaces_subinterface_state_pkts{index=\"123\",name=\"xe-0/0/0\",some=\"label\"} 232323\n",
		},
		{
			name: "Test #5",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "foo,bar=baz",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/1']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{bar=\"baz\",name=\"xe-0/0/0\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{name=\"xe-0/0/1\"} 1337\n",
		},
		{
			name: "Test #6",
			config: &config.Config{
				StringValueMapping: map[string]map[string]int{
					"/interfaces/interface/state/oper-state": map[string]int{
						"UP":   100,
						"DOWN": 200,
					},
				},
				Targets: []*config.Target{
					{
						Paths: []*config.Path{
							{
								Path: "/interfaces/",
							},
						},
					},
				},
			},
			testdata: []pb.OpenConfigData{
				{
					Kv: []*pb.KeyValue{
						{
							Key: "__prefix__",
							Value: &pb.KeyValue_StrValue{
								StrValue: "/interfaces/",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/description",
							Value: &pb.KeyValue_StrValue{
								StrValue: "foo,bar=baz",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/oper-state",
							Value: &pb.KeyValue_StrValue{
								StrValue: "UP",
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/some-double",
							Value: &pb.KeyValue_DoubleValue{
								DoubleValue: 1338,
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/some-uint",
							Value: &pb.KeyValue_UintValue{
								UintValue: 232323,
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/some-sint",
							Value: &pb.KeyValue_SintValue{
								SintValue: 4242,
							},
						},
						{
							Key: "interface[name='xe-0/0/0']/state/some-bool",
							Value: &pb.KeyValue_BoolValue{
								BoolValue: true,
							},
						},
						{
							Key: "interface[name='xe-0/0/1']/state/pkts",
							Value: &pb.KeyValue_IntValue{
								IntValue: 1337,
							},
						},
					},
				},
			},
			expected: "# HELP interfaces_interface_state_oper_state interfaces/interface/state/oper-state\n# TYPE interfaces_interface_state_oper_state gauge\ninterfaces_interface_state_oper_state{bar=\"baz\",name=\"xe-0/0/0\"} 100\n# HELP interfaces_interface_state_pkts interfaces/interface/state/pkts\n# TYPE interfaces_interface_state_pkts gauge\ninterfaces_interface_state_pkts{name=\"xe-0/0/1\"} 1337\n# HELP interfaces_interface_state_some_bool interfaces/interface/state/some-bool\n# TYPE interfaces_interface_state_some_bool gauge\ninterfaces_interface_state_some_bool{bar=\"baz\",name=\"xe-0/0/0\"} 1\n# HELP interfaces_interface_state_some_double interfaces/interface/state/some-double\n# TYPE interfaces_interface_state_some_double gauge\ninterfaces_interface_state_some_double{bar=\"baz\",name=\"xe-0/0/0\"} 1338\n# HELP interfaces_interface_state_some_sint interfaces/interface/state/some-sint\n# TYPE interfaces_interface_state_some_sint gauge\ninterfaces_interface_state_some_sint{bar=\"baz\",name=\"xe-0/0/0\"} 4242\n# HELP interfaces_interface_state_some_uint interfaces/interface/state/some-uint\n# TYPE interfaces_interface_state_some_uint gauge\ninterfaces_interface_state_some_uint{bar=\"baz\",name=\"xe-0/0/0\"} 232323\n",
		},
	}

	for _, test := range tests {
		test.config.LoadDefaults()
		collector := New(test.config)

		bufSize := 1024 * 1024
		lis := bufconn.Listen(bufSize)
		s := grpc.NewServer()
		telemetryServer := &mockTelemetryServer{
			testdata: test.testdata,
		}
		telemetryServer.wg.Add(1)
		pb.RegisterOpenConfigTelemetryServer(s, telemetryServer)
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()

		var serveWG sync.WaitGroup
		for _, confTarget := range test.config.Targets {
			serveWG.Add(1)
			go func(confTarget *config.Target) {
				ta := collector.AddTarget(confTarget, test.config.StringValueMapping)

				ctx := context.Background()
				conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
					return lis.Dial()
				}), grpc.WithInsecure())
				if err != nil {
					t.Fatalf("Failed to dial bufnet: %v", err)
				}
				defer conn.Close()

				ta.maxReads = len(test.testdata)
				ta.Serve(conn)
				serveWG.Done()
			}(confTarget)
		}

		telemetryServer.wg.Wait() // Wait for grpc server to write all data
		serveWG.Wait()            // Wait for target server to save all data in the tree

		//time.Sleep(time.Second * 5)
		w := newMockHTTPResponseWriter()
		r := &http.Request{
			Method: "GET",
			URL: &url.URL{
				Path: "/metrics",
			},
		}

		reg := prometheus.NewRegistry()
		reg.MustRegister(collector)

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)

		assert.Equal(t, test.expected, string(w.buf.Bytes()), test.name)

		//telemetryServer.stop <- struct{}{}
	}
}

type mockHTTPResponseWriter struct {
	buf *bytes.Buffer
}

func newMockHTTPResponseWriter() *mockHTTPResponseWriter {
	return &mockHTTPResponseWriter{
		buf: bytes.NewBuffer(nil),
	}
}

func (m *mockHTTPResponseWriter) WriteHeader(int) {

}

func (m *mockHTTPResponseWriter) Write(input []byte) (int, error) {
	m.buf.Write(input)
	return len(input), nil
}

func (m *mockHTTPResponseWriter) Header() http.Header {
	return http.Header{}
}

func TestExtractLabels(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedPrefix string
		expectedLabels []label
		wantFail       bool
	}{
		{
			name:           "Test #1",
			input:          "/interfaces/interface[name='xe-0/0/34:0']/",
			expectedPrefix: "/interfaces/interface/",
			expectedLabels: []label{
				{
					key:   "name",
					value: "xe-0/0/34:0",
				},
			},
		},
		{
			name:           "Test #2",
			input:          "state/counters/out-queue[queue-number=0]/bytes/",
			expectedPrefix: "state/counters/out-queue/bytes/",
			expectedLabels: []label{
				{
					key:   "queue-number",
					value: "0",
				},
			},
		},
		{
			name:           "Test #3",
			input:          "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/neighbors/neighbor[neighbor-address='1.2.3.4']/",
			expectedPrefix: "/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/",
			expectedLabels: []label{
				{
					key:   "instance-name",
					value: "master",
				},
				{
					key:   "neighbor-address",
					value: "1.2.3.4",
				},
			},
		},
	}

	for _, test := range tests {
		prefix, labels, err := parseKey(test.input)
		if test.wantFail {
			if err != nil {
				continue
			}

			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expectedPrefix, prefix, test.name)
		assert.Equal(t, test.expectedLabels, labels, test.name)
	}
}
