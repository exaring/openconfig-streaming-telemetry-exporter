package collector

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	descLabelRegexp    = regexp.MustCompile("([a-zA-Z_][a-zA-Z0-9_]*)")
	metricNameReplacer = strings.NewReplacer(
		"/", "_",
		"-", "_",
		"'", "",
	)
	labelValueReplacer = strings.NewReplacer(
		"'", "",
	)
	labelKeyReplacer = strings.NewReplacer(
		"-", "_",
		"'", "",
	)
)

type tree struct {
	lock sync.RWMutex
	root *node
}

type node struct {
	id          identifier
	real        bool
	value       interface{}
	description string
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

func (t *tree) dump() []string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.root == nil {
		return nil
	}

	return t.root.dump(0)
}

func (t *tree) setDescription(path string, v string) {
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

	res := newMetricSet()
	t.root.getMetrics("", res, []label{}, []label{})

	return res.get()
}

func (n *node) dump(level int) []string {
	ret := make([]string, 0)

	ret = append(ret, "|\n")
	ret = append(ret, fmt.Sprintf("%s[%s](%v) = %v\n", n.id.name, n.id.labels, n.description, n.value))

	for _, c := range n.children {
		ret = append(ret, c.dump(level+1)...)
	}

	for i := range ret {
		for j := 0; j < level; j++ {
			ret[i] = " " + ret[i]
		}
		ret[i] = "|" + ret[i]
	}

	return ret
}

func (n *node) setDescription(path []identifier, v string) {
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
	parts := strings.Split(n.description, ",")
	for _, p := range parts {
		keyValue := strings.Split(p, "=")
		if len(keyValue) != 2 {
			return []string{}
		}

		if !descLabelRegexp.Match([]byte(keyValue[0])) {
			return []string{}
		}
	}

	return parts
}

func (n *node) getMetrics(path string, res *metricSet, labels []label, descriptionLabels []label) {
	if path == "" {
		path = n.id.name
	} else {
		path = path + "/" + n.id.name
	}

	if n.id.labels != "" {
		newLabels := labelStringToLabels(n.id.labels)
		mergedLabels := make([]label, len(labels)+len(newLabels))
		for i, label := range labels {
			mergedLabels[i] = label
		}

		for i, label := range newLabels {
			mergedLabels[len(labels)+i] = label
		}

		labels = mergedLabels
	}

	if n.description != "" {
		descriptionLabels = labelStringToLabels(n.description)
	}

	keys := make([]identifier, len(n.children))
	i := 0
	for key := range n.children {
		keys[i] = key
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].name < keys[j].name
	})

	if n.real {
		m := metric{
			name:   path,
			value:  n.value,
			labels: append(labels, descriptionLabels...),
		}

		res.append(m)
	}

	for _, key := range keys {
		n.children[key].getMetrics(path, res, labels, descriptionLabels)
	}
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

func labelStringToLabels(input string) []label {
	res := make([]label, 0, 10)
	for _, labelStr := range strings.Split(input, ",") {
		kv := strings.Split(labelStr, "=")
		if len(kv) != 2 {
			continue
		}

		if !descLabelRegexp.Match([]byte(kv[0])) {
			continue
		}

		res = append(res, label{
			key:   labelKeyReplacer.Replace(kv[0]),
			value: labelValueReplacer.Replace(kv[1]),
		})
	}

	return res
}

func dropSlashPrefixSuffix(p string) []rune {
	start := 0
	end := len(p)

	if strings.HasPrefix(p, "/") {
		start = 1
	}

	if strings.HasSuffix(p, "/") {
		end = len(p) - 1
	}

	return []rune(p)[start:end]
}

func slashCount(runes []rune) int {
	count := 0
	for _, r := range runes {
		if r == '/' {
			count++
		}
	}

	return count
}

func pathToIdentifiers(p string) []identifier {
	runes := dropSlashPrefixSuffix(p)
	res := make([]identifier, 0, slashCount(runes))

	bracesLevel := 0
	tmp := make([]rune, 0, 256)
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
				res = append(res, pathElementToIdentifier(tmp))
				tmp = tmp[:0]

				/*tmpCopy := make([]rune, len(tmp))
				copy(tmpCopy, tmp)
				res = append(res, tmpCopy)
				tmp = tmp[:0] // clear tmp instead of allocating a new slice*/
				continue
			}
		}

		tmp = append(tmp, runes[i])
	}

	if len(tmp) > 0 {
		//res = append(res, tmp)
		res = append(res, pathElementToIdentifier(tmp))
	}

	return res
}

func pathElementToIdentifier(data []rune) identifier {
	key := make([]rune, 0, 25)
	labelsString := ""
	withinAngledBraces := false
	tmp := make([]rune, 0, 25)

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
