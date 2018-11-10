package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlashCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Test #1",
			input:    "/interfaces/foo/bar/",
			expected: 4,
		},
	}

	for _, test := range tests {
		x := slashCount([]rune(test.input))
		assert.Equal(t, test.expected, x, test.name)
	}
}
func TestGetMetrics(t *testing.T) {
	tests := []struct {
		name  string
		input []struct {
			path  string
			value float64
		}
		expected []metric
	}{
		{
			name: "Test #1",
			input: []struct {
				path  string
				value float64
			}{
				{
					path:  "/interfaces/",
					value: 100,
				},
			},
			expected: []metric{
				{
					name:   "interfaces",
					value:  float64(100),
					labels: []label{},
				},
			},
		},
		{
			name: "Test #2",
			input: []struct {
				path  string
				value float64
			}{
				{
					path:  "/interfaces[foo='bar']/bgp/something[some='label']/",
					value: 200,
				},
			},
			expected: []metric{
				{
					name:  "interfaces/bgp/something",
					value: float64(200),
					labels: []label{
						{
							key:   "foo",
							value: "bar",
						},
						{
							key:   "some",
							value: "label",
						},
					},
				},
			},
		},
		{
			name: "Test #3",
			input: []struct {
				path  string
				value float64
			}{
				{
					path:  "/interfaces[foo='bar']/bgp/something[some='label']/",
					value: 200,
				},
				{
					path:  "/interfaces[foo='bar']/bgp/something[some='crap']/",
					value: 300,
				},
			},
			expected: []metric{
				{
					name:  "interfaces/bgp/something",
					value: float64(200),
					labels: []label{
						{
							key:   "foo",
							value: "bar",
						},
						{
							key:   "some",
							value: "label",
						},
					},
				},
				{
					name:  "interfaces/bgp/something",
					value: float64(300),
					labels: []label{
						{
							key:   "foo",
							value: "bar",
						},
						{
							key:   "some",
							value: "crap",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		tr := newTree()
		for _, input := range test.input {
			tr.insert(input.path, input.value)
		}

		m := tr.getMetrics()
		assert.Equal(t, test.expected, m, test.name)
	}
}

func TestTokenizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []identifier
	}{
		{
			name:  "Test #1",
			input: "/interfaces/interface[name='xe-0/0/0']/pkts/",
			expected: []identifier{
				{
					name: "interfaces",
				},
				{
					name:   "interface",
					labels: "name='xe-0/0/0'",
				},
				{
					name: "pkts",
				},
			},
		},
	}

	for _, test := range tests {
		tokens := tokenizePath(test.input)
		assert.Equal(t, test.expected, tokens, test.name)
	}
}

func TestPathToIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []identifier
	}{
		{
			name:  "Test #1",
			input: "/interfaces/interface[name='xe-0/0/0']/pkts/",
			expected: []identifier{
				{
					name: "interfaces",
				},
				{
					name:   "interface",
					labels: "name='xe-0/0/0'",
				},
				{
					name: "pkts",
				},
			},
		},
		{
			name:  "Test #2",
			input: "/interfaces/interface[name='xe-0/0/0']/pkts/state[with='label']",
			expected: []identifier{
				{
					name: "interfaces",
				},
				{
					name:   "interface",
					labels: "name='xe-0/0/0'",
				},
				{
					name: "pkts",
				},
				{
					name:   "state",
					labels: "with='label'",
				},
			},
		},
	}

	for _, test := range tests {
		ids := pathToIdentifiers(test.input)
		assert.Equal(t, test.expected, ids, test.name)
	}
}

func TestPathElementToIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected identifier
	}{
		{
			name:  "Test #1",
			input: "interfaces",
			expected: identifier{
				name: "interfaces",
			},
		},
		{
			name:  "Test #1",
			input: "interfaces[name='xe-0/0/0']",
			expected: identifier{
				name:   "interfaces",
				labels: "name='xe-0/0/0'",
			},
		},
	}

	for _, test := range tests {
		id := pathElementToIdentifier([]rune(test.input))
		assert.Equal(t, test.expected, id, test.name)
	}
}
