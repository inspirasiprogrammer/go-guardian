package store

import (
	"container/list"
	"net/http"
	"sync"
	"time"
)

// LRU implements a fixed-size thread safe LRU cache.
// It is based on the LRU cache in Groupcache.
type LRU struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted OnEvicted

	// TTL To expire a value in cache.
	// 0 TTL means no expiry policy specified.
	TTL time.Duration

	MU *sync.Mutex

	c *core
}

// New creates a new LRU Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxEntries int) *LRU {
	return &LRU{
		MaxEntries: maxEntries,
		MU:         new(sync.Mutex),
		c: &core{
			ll:    list.New(),
			cache: make(map[string]*list.Element),
		},
	}
}

// Store sets the value for a key.
func (l *LRU) Store(key string, value interface{}, _ *http.Request) error {
	l.MU.Lock()
	defer l.MU.Unlock()
	l.set()

	e := l.c.store(key, value)
	l.c.ll.MoveToFront(e)

	if l.MaxEntries != 0 && l.c.ll.Len() > l.MaxEntries {
		l.removeOldest()
	}

	return nil
}

func (l *LRU) set() {
	l.c.onEvicted = l.OnEvicted
	l.c.ttl = l.TTL
}

func (l *LRU) withTTL(r *record) {
	if l.TTL > 0 {
		r.Exp = time.Now().UTC().Add(l.TTL)
	}
}

// Load returns the value stored in the Cache for a key, or nil if no value is present.
// The ok result indicates whether value was found in the Cache.
func (l *LRU) Load(key string, _ *http.Request) (interface{}, bool, error) {
	l.MU.Lock()
	defer l.MU.Unlock()

	if l.c == nil {
		return nil, false, nil
	}

	e, ok, err := l.c.load(key)

	if ok && err == nil {
		l.c.ll.MoveToFront(e)
		return e.Value.(*record).Value, ok, err
	}

	return nil, ok, err
}

// Delete the value for a key.
func (l *LRU) Delete(key string, _ *http.Request) error {
	l.MU.Lock()
	defer l.MU.Unlock()

	if l.c == nil {
		return nil
	}

	l.c.delete(key)

	return nil
}

// RemoveOldest removes the oldest item from the cache.
func (l *LRU) RemoveOldest() {
	l.MU.Lock()
	defer l.MU.Unlock()
	l.removeOldest()
}

func (l *LRU) removeOldest() {
	if l.c == nil {
		return
	}

	if ele := l.c.ll.Back(); ele != nil {
		l.c.removeElement(ele)
	}
}


// Len returns the number of items in the cache.
func (l *LRU) Len() int {
	l.MU.Lock()
	defer l.MU.Unlock()

	if l.c == nil {
		return 0
	}

	return l.c.ll.Len()
}

// Clear purges all stored items from the cache.
func (l *LRU) Clear() {
	l.MU.Lock()
	defer l.MU.Unlock()

	if l.OnEvicted != nil {
		for _, e := range l.c.cache {
			kv := e.Value.(*record)
			l.OnEvicted(kv.Key, kv.Value)
		}
	}
}

// Keys return cache records keys.
func (l *LRU) Keys() []string {
	l.MU.Lock()
	defer l.MU.Unlock()

	return l.c.keys()
}
