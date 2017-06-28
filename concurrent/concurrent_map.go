package concurrent

import (
	"sync"
)

type ConcurrentMap interface {
	AddOrUpdate(string, interface{}, func(string, interface{}) interface{})
	Delete(string)
	Get(string) (interface{}, bool)
	GetKeys() []string
	GetValues() []interface{}
	IsEmpty() bool
}

type concurrentMap struct {
	cmap map[string]interface{}
	mu   sync.RWMutex
}

func NewConcurrentMap() ConcurrentMap {
	return &concurrentMap{cmap: make(map[string]interface{})}
}

func (c *concurrentMap) AddOrUpdate(key string, value interface{}, updValFactory func(key string, oldVal interface{}) interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currVal, ok := c.cmap[key]

	if !ok {
		c.cmap[key] = value
	} else {
		c.cmap[key] = updValFactory(key, currVal)
	}
}

func (c *concurrentMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cmap, key)
}

func (c *concurrentMap) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := []string{}

	for k, _ := range c.cmap {
		keys = append(keys, k)
	}
	return keys
}

func (c *concurrentMap) GetValues() []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := []interface{}{}

	for _, v := range c.cmap {
		values = append(values, v)
	}
	return values

}

func (c *concurrentMap) IsEmpty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	isEmpety := len(c.cmap) == 0
	return isEmpety
}

func (c *concurrentMap) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.cmap[key]
	return value, ok
}
