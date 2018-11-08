package collector

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	pb "github.com/exaring/openconfig-streaming-telemetry-exporter/pkg/telemetry"
)

var (
	descLabelRegexp = regexp.MustCompile("([a-zA-Z_][a-zA-Z0-9_]*=.*)(,[a-zA-Z_][a-zA-Z0-9_]*=.*)*")
)

type tree struct {
	lock sync.RWMutex
	root *node
}

type node struct {
	id          identifier
	real        bool
	value       interface{}
	description interface{}
	children    map[identifier]*node
}

type identifier struct {
	name   string
	labels string
}

func newTree() *tree {
	return &tree{}
}

func newNode(id identifier) *node {
	return &node{
		id:       id,
		children: make(map[identifier]*node),
	}
}

func (t *tree) setDescription(path string, v interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	ids := pathToIdentifiers(path)
	if t.root == nil {
		t.root = newNode(identifier{})
	}

	t.root.setDescription(ids, v)
}

func (t *tree) insert(path string, v interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	ids := pathToIdentifiers(path)

	if t.root == nil {
		t.root = newNode(identifier{})
	}

	t.root.insert(ids, v)
}

func (t *tree) getMetrics() []metric {
	if t.root == nil {
		return nil
	}

	return t.root.getMetrics()
}

func (n *node) setDescription(path []identifier, v interface{}) {
	if len(path) > 0 {
		if _, ok := n.children[path[0]]; !ok {
			n.children[path[0]] = newNode(path[0])
		}

		n.children[path[0]].setDescription(path[1:], v)
		return
	}

	n.description = v
}

func (n *node) descLabels() []string {
	switch v := n.description.(type) {
	case *pb.KeyValue_StrValue:
		if descLabelRegexp.Match([]byte(v.StrValue)) {
			return strings.Split(v.StrValue, ",")
		}
	}

	return []string{}
}

func (n *node) getMetrics() []metric {
	res := make([]metric, 0)

	keys := make([]identifier, len(n.children))
	i := 0
	for key := range n.children {
		keys[i] = key
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].name < keys[j].name
	})

	for _, key := range keys {
		for _, m := range n.children[key].getMetrics() {
			if n.id.name != "" {
				m.name = n.id.name + "/" + m.name
			}

			if n.id.labels != "" {
				m.labels = append(m.labels, strings.Split(n.id.labels, ",")...)
			}

			if n.description != "" {
				m.labels = append(m.labels, n.descLabels()...)
			}

			res = append(res, m)
		}
	}

	if n.real {
		m := metric{
			name:  n.id.name,
			value: n.value,
		}

		if n.id.labels != "" {
			m.labels = strings.Split(n.id.labels, ",")
		}

		res = append(res, m)
	}

	return res
}

func (n *node) insert(path []identifier, v interface{}) {
	if len(path) > 0 {
		if _, ok := n.children[path[0]]; !ok {
			n.children[path[0]] = newNode(path[0])
		}

		n.children[path[0]].insert(path[1:], v)
		return
	}

	n.real = true
	n.value = v
}

func pathToIdentifiers(p string) []identifier {
	tokens := tokenizePath(p)
	ids := make([]identifier, len(tokens))

	for i, t := range tokens {
		ids[i] = pathElementToIdentifier(t)
	}

	return ids
}

func dropSlashPrefixSuffix(p string) string {
	if strings.HasPrefix(p, "/") {
		p = string([]rune(p)[1:])
	}

	if strings.HasSuffix(p, "/") {
		p = string([]rune(p)[:len(p)-1])
	}

	return p
}

func tokenizePath(p string) []string {
	p = dropSlashPrefixSuffix(p)
	runes := []rune(p)
	res := make([]string, 0, 10)

	bracesLevel := 0
	tmp := make([]rune, 0, 15)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '[' {
			bracesLevel++
		}

		if bracesLevel > 0 {
			if runes[i] == ']' {
				bracesLevel--
			}
		}

		if runes[i] == '/' {
			if bracesLevel == 0 {
				res = append(res, string(tmp))
				tmp = make([]rune, 0, 15)
				continue
			}
		}

		tmp = append(tmp, runes[i])
	}

	if len(tmp) > 0 {
		res = append(res, string(tmp))
	}

	return res
}

func pathElementToIdentifier(e string) identifier {
	data := []rune(e)

	key := make([]rune, 0, 15)
	labelsString := ""
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
			labelsString = string(tmp)
			withinAngledBraces = false
			tmp = make([]rune, 0)
			continue
		}

		tmp = append(tmp, data[i])
	}

	return identifier{
		name:   string(key),
		labels: labelsString,
	}
}
