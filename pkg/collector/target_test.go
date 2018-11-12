package collector

import (
	"testing"

	pb "github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestProcessOpenConfigData(t *testing.T) {
	tests := []struct {
		name     string
		target   *Target
		input    *pb.OpenConfigData
		expected *Target
	}{
		{
			name: "All ok",
			target: &Target{
				metrics: newTree(),
			},
			input: &pb.OpenConfigData{
				Kv: []*pb.KeyValue{
					{
						Key: "__prefix__",
						Value: &pb.KeyValue_StrValue{
							StrValue: "foobar/",
						},
					},
					{
						Key: "baz",
						Value: &pb.KeyValue_StrValue{
							StrValue: "hello world",
						},
					},
				},
			},
			expected: &Target{
				metrics: &tree{
					root: &node{
						id: identifier{},
						children: []node{
							{
								id: identifier{
									name: "foobar",
								},
								children: []node{
									{
										id: identifier{
											name: "baz",
										},
										real: true,
										value: &pb.KeyValue_StrValue{
											StrValue: "hello world",
										},
										children: []node{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Ignore __",
			target: &Target{
				metrics: newTree(),
			},
			input: &pb.OpenConfigData{
				Kv: []*pb.KeyValue{
					{
						Key: "__prefix__",
						Value: &pb.KeyValue_StrValue{
							StrValue: "foobar/",
						},
					},
					{
						Key: "__some_shit",
						Value: &pb.KeyValue_StrValue{
							StrValue: "blaah/",
						},
					},
					{
						Key: "baz",
						Value: &pb.KeyValue_StrValue{
							StrValue: "hello world",
						},
					},
				},
			},
			expected: &Target{
				metrics: &tree{
					root: &node{
						id: identifier{},
						children: []node{
							{
								id: identifier{
									name: "foobar",
								},
								children: []node{
									{
										id: identifier{
											name: "baz",
										},
										real: true,
										value: &pb.KeyValue_StrValue{
											StrValue: "hello world",
										},
										children: []node{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "prefix with nil value",
			target: &Target{
				metrics: newTree(),
			},
			input: &pb.OpenConfigData{
				Kv: []*pb.KeyValue{
					{
						Key:   "__prefix__",
						Value: nil,
					},
					{
						Key: "baz",
						Value: &pb.KeyValue_StrValue{
							StrValue: "hello world",
						},
					},
				},
			},
			expected: &Target{
				metrics: &tree{
					root: &node{
						id: identifier{},
						children: []node{
							{
								id: identifier{
									name: "baz",
								},
								real: true,
								value: &pb.KeyValue_StrValue{
									StrValue: "hello world",
								},
								children: []node{},
							},
						},
					},
				},
			},
		},
		{
			name: "non-string prefix",
			target: &Target{
				metrics: newTree(),
			},
			input: &pb.OpenConfigData{
				Kv: []*pb.KeyValue{
					{
						Key: "__prefix__",
						Value: &pb.KeyValue_IntValue{
							IntValue: 1337,
						},
					},
					{
						Key: "baz",
						Value: &pb.KeyValue_StrValue{
							StrValue: "hello world",
						},
					},
				},
			},
			expected: &Target{
				metrics: &tree{
					root: &node{
						id: identifier{},
						children: []node{
							{
								id: identifier{
									name: "baz",
								},
								real: true,
								value: &pb.KeyValue_StrValue{
									StrValue: "hello world",
								},
								children: []node{},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.target.processOpenConfigData(test.input)
		assert.Equal(t, test.expected, test.target, test.name)
	}
}
