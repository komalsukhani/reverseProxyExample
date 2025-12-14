package internal

import (
	"net/http"
	"sync"
	"time"
)

type MemoryCache struct {
	mu      sync.RWMutex
	records map[string]*Record
	ttl     time.Duration
}

type Record struct {
	StatusCode int //if in future need to cache other status codes
	Body       []byte
	Headers    http.Header
	expiry     time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		records: make(map[string]*Record),
		ttl:     ttl,
	}
}

func (cache *MemoryCache) Get(k string) *Record {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	r, ok := cache.records[k]
	if !ok {
		return nil
	}

	//If the record has expired, delete it from the cache
	if time.Now().After(r.expiry) {
		delete(cache.records, k)
		return nil
	}

	return r
}

func (cache *MemoryCache) Set(k string, data *Record) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	data.expiry = time.Now().Add(cache.ttl)

	cache.records[k] = data
}

func (cache *MemoryCache) Count() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return len(cache.records)
}
