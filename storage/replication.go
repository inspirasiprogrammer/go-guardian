package storage

import "net/http"

// ReplicationCache implements cache and holding two caches instances to replicate data between them on load and store data,
// ReplicationCache	is typically used to replicate data between in-memory and a persistent caches,
// to obtain data consistency and persistency between runs of a program or scaling purposes.
// Most users will use distributed caching instead.
type ReplicationCache struct {
	InMemory   Cache
	Persistent Cache
}

func (r *ReplicationCache) Load(key string, req *http.Request) (interface{}, bool, error) {
	return nil, false, nil
}

func (r *ReplicationCache) Store(key string, value interface{}, req *http.Request) error { return nil }