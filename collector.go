package relay

//Collector defines a typ of map string
type Collector map[string]interface{}

//StringEachfunc defines the type of the Mappable.Each rule
type StringEachfunc func(interface{}, string, func())

//NewCollector returns a new collector instance
func NewCollector() Collector {
	return make(Collector)
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

//Clear clears the collector
func (c Collector) Clear() {
	for k := range c {
		delete(c, k)
	}
}
