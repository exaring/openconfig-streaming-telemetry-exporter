package collector

import "sync"

type idCache struct {
	cache   map[string][]identifier
	cacheMu sync.RWMutex
}

func newIDCache() *idCache {
	return &idCache{
		cache: make(map[string][]identifier),
	}
}

func (c *idCache) lookup(p string) []identifier {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	return c.cache[p]
}

func (c *idCache) set(p string, ids []identifier) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[p] = ids
}
