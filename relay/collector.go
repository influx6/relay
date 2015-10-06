package relay

import "sync"

//StringEachfunc defines the type of the Mappable.Each rule
type StringEachfunc func(interface{}, string, func())

//Collector defines a typ of map string
type Collector map[string]interface{}

//NewCollector returns a new collector instance
func NewCollector() Collector {
	return make(Collector)
}

//ToMap makes a new clone of this map[string]interface{}
func (c Collector) ToMap() map[string]interface{} {
	return c.Clone()
}

//Clone makes a new clone of this collector
func (c Collector) Clone() Collector {
	col := make(Collector)
	col.Copy(c)
	return col
}

//Remove deletes a key:value pair
func (c Collector) Remove(k string) {
	if c.Has(k) {
		delete(c, k)
	}
}

//Keys return the keys of the Collector
func (c Collector) Keys() []string {
	var keys []string
	c.Each(func(_ interface{}, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

//Get returns the value with the key
func (c Collector) Get(k string) interface{} {
	return c[k]
}

//Has returns if a key exists
func (c Collector) Has(k string) bool {
	_, ok := c[k]
	return ok
}

//HasMatch checks if key and value exists and are matching
func (c Collector) HasMatch(k string, v interface{}) bool {
	if c.Has(k) {
		return c.Get(k) == v
	}
	return false
}

//Set puts a specific key:value into the collector
func (c Collector) Set(k string, v interface{}) {
	c[k] = v
}

//Copy copies the map into the collector
func (c Collector) Copy(m map[string]interface{}) {
	for v, k := range m {
		c.Set(v, k)
	}
}

//Each iterates through all items in the collector
func (c Collector) Each(fx StringEachfunc) {
	var state bool
	for k, v := range c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
}

//Clear clears the collector
func (c Collector) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// SyncCollector provides a mutex controlled map
type SyncCollector struct {
	c  Collector
	rw sync.RWMutex
}

//NewSyncCollector returns a new collector instance
func NewSyncCollector() *SyncCollector {
	so := SyncCollector{c: make(Collector)}
	return &so
}

//Clone makes a new clone of this collector
func (c *SyncCollector) Clone() *SyncCollector {
	var co Collector

	c.rw.RLock()
	co = c.c.Clone()
	c.rw.RUnlock()

	so := SyncCollector{c: co}
	return &so
}

//ToMap makes a new clone of this map[string]interface{}
func (c *SyncCollector) ToMap() map[string]interface{} {
	var co Collector
	c.rw.RLock()
	co = c.c.Clone()
	c.rw.RUnlock()
	return co
}

//Remove deletes a key:value pair
func (c *SyncCollector) Remove(k string) {
	c.rw.Lock()
	c.c.Remove(k)
	c.rw.Unlock()
}

//Set puts a specific key:value into the collector
func (c *SyncCollector) Set(k string, v interface{}) {
	c.rw.Lock()
	c.c.Set(k, v)
	c.rw.Unlock()
}

//Copy copies the map into the collector
func (c *SyncCollector) Copy(m map[string]interface{}) {
	for v, k := range m {
		c.Set(v, k)
	}
}

//Each iterates through all items in the collector
func (c *SyncCollector) Each(fx StringEachfunc) {
	var state bool
	c.rw.RLock()
	for k, v := range c.c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
	c.rw.RUnlock()
}

//Keys return the keys of the Collector
func (c *SyncCollector) Keys() []string {
	var keys []string
	c.Each(func(_ interface{}, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

//Get returns the value with the key
func (c *SyncCollector) Get(k string) interface{} {
	var v interface{}
	c.rw.RLock()
	v = c.c.Get(k)
	c.rw.RUnlock()
	return v
}

//Has returns if a key exists
func (c *SyncCollector) Has(k string) bool {
	var ok bool
	c.rw.RLock()
	_, ok = c.c[k]
	c.rw.RUnlock()
	return ok
}

//HasMatch checks if key and value exists and are matching
func (c *SyncCollector) HasMatch(k string, v interface{}) bool {
	// c.rw.RLock()
	// defer c.rw.RUnlock()
	if c.Has(k) {
		return c.Get(k) == v
	}
	return false
}

//Clear clears the collector
func (c *SyncCollector) Clear() {
	for k := range c.c {
		c.rw.Lock()
		delete(c.c, k)
		c.rw.Unlock()
	}
}
