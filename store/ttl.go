package store

import (
	"container/list"
	"time"
)

type record struct {
	Exp   time.Time
	timer *time.Timer
	Key   string
	Value interface{}
}

type cache struct {
	list      *list.List
	items     map[string]*list.Element
	ttl       time.Duration
	onEvicted OnEvicted
}

func (c *cache) store(key string, value interface{}) *list.Element {
	if e, ok := c.items[key]; ok {
		c.removeElement(e)
	}

	r := &record{Key: key, Value: value}
	r = c.withTTL(r, c.ttl)
	e := c.list.PushBack(r)
	c.items[key] = e
	return e
}

func (c *cache) update(key string, value interface{}) {
	if e, ok := c.items[key]; ok {
		r := e.Value.(*record)
		r.Value = value
	}
}

func (c *cache) load(key string) (*list.Element, bool, error) {

	if e, ok := c.items[key]; ok {
		r := e.Value.(*record)

		if c.ttl > 0 {
			if time.Now().UTC().After(r.Exp) {
				c.evict(e)
				return e, ok, ErrCachedExp
			}
		}

		return e, ok, nil
	}

	return nil, false, nil
}

func (c *cache) delete(key string) {
	if e, ok := c.items[key]; ok {
		c.evict(e)
	}
}

func (c *cache) removeElement(e *list.Element) {
	c.list.Remove(e)
	r := e.Value.(*record)
	if r.timer != nil {
		r.timer.Stop()
	}
	delete(c.items, r.Key)
}

func (c *cache) evict(e *list.Element) {
	c.removeElement(e)
	r := e.Value.(*record)
	if c.onEvicted != nil {
		c.onEvicted(r.Key, r.Value)
	}
}

func (c *cache) keys() (keys []string) {
	for k := range c.items {
		keys = append(keys, k)
	}

	return keys
}

func (c *cache) len() int {
	return c.list.Len()
}

func (c *cache) clear() {
	if c.onEvicted == nil {
		return
	}

	for _, e := range c.items {
		c.evict(e)
	}
}

func (c *cache) withTTL(r *record, ttl time.Duration) *record {
	if ttl > 0 {
		r.timer = time.AfterFunc(ttl, func() {
			c.load(r.Key)
		})

		r.Exp = time.Now().UTC().Add(ttl)
	}

	return r
}

func newCache() *cache {
	c := new(cache)
	c.list = list.New()
	c.items = make(map[string]*list.Element)
	return c
}
