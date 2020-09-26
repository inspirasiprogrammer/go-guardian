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

type core struct {
	ll        *list.List
	cache     map[string]*list.Element
	ttl       time.Duration
	onEvicted OnEvicted
}

func (c *core) store(key string, value interface{}) *list.Element {
	if e, ok := c.cache[key]; ok {
		r := e.Value.(*record)
		r.Value = value
		e.Value = c.withTTL(r, c.ttl)
		return e
	}

	r := &record{Key: key, Value: value}
	r = c.withTTL(r, c.ttl)
	e := c.ll.PushBack(r)
	c.cache[key] = e
	return e
}

func (c *core) load(key string) (*list.Element, bool, error) {

	if e, ok := c.cache[key]; ok {
		r := e.Value.(*record)

		if c.ttl > 0 {
			if time.Now().UTC().After(r.Exp) {
				c.removeElement(e)
				return e, ok, ErrCachedExp
			}
		}

		return e, ok, nil
	}

	return nil, false, nil
}

func (c *core) delete(key string) {
	if e, ok := c.cache[key]; ok {
		c.removeElement(e)
	}
}

func (c *core) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*record)
	delete(c.cache, kv.Key)
	if c.onEvicted != nil {
		c.onEvicted(kv.Key, kv.Value)
	}
}

func (c *core) keys() (keys []string) {
	for k := range c.cache {
		keys = append(keys, k)
	}

	return keys
}

func (c *core) withTTL(r *record, ttl time.Duration) *record {
	if r.timer != nil {
		r.timer.Stop()
	}

	if ttl > 0 {
		r.timer = time.AfterFunc(ttl, func() {
			c.load(r.Key)
		})

		r.Exp = time.Now().UTC().Add(ttl)
	}

	return r
}
