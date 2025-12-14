package internal

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

var ErrCacheFull = errors.New("memory cache is full")

// TODO: Remove expired entries if size is full
type MemoryCache struct {
	mu                sync.RWMutex
	records           map[string]*Record
	ttl               time.Duration
	maxSize           int
	remainingCapacity int
}

type Record struct {
	StatusCode int //if in future need to cache other status codes
	Body       []byte
	Headers    http.Header
	expiry     time.Time
	size       int
}

func NewMemoryCache(ttl time.Duration, maxSize int) *MemoryCache {
	return &MemoryCache{
		records:           make(map[string]*Record),
		ttl:               ttl,
		maxSize:           maxSize,
		remainingCapacity: maxSize,
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
		cache.remainingCapacity += r.size
		delete(cache.records, k)

		return nil
	}

	return r
}

func (cache *MemoryCache) Set(k string, data *Record) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	data.size = len(data.Body)

	existingRecord, ok := cache.records[k]
	if !ok {
		//skipping entry in cache as cache has reached it's limit
		if data.size > cache.remainingCapacity {
			return ErrCacheFull
		}

		cache.remainingCapacity -= data.size
	} else {
		oldSize := existingRecord.size

		//skipping entry in cache as cache has reached it's limit
		if cache.remainingCapacity+data.size-oldSize > cache.maxSize {
			return ErrCacheFull
		}

		cache.remainingCapacity = +data.size - oldSize
	}

	data.expiry = time.Now().Add(cache.ttl)

	cache.records[k] = data

	return nil
}

func (cache *MemoryCache) Count() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return len(cache.records)
}
