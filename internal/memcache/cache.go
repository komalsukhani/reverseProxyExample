package memcache

import (
	"container/list"
	"errors"
	"net/http"
	"sync"
	"time"
)

var ErrMaxRecordSizeExceed = errors.New("request exceeds max record size")

type MemoryCache struct {
	mu                sync.RWMutex
	records           map[string]*Record
	ttl               time.Duration
	maxCacheSize      int
	maxRecordSize     int
	remainingCapacity int

	ll *list.List
}

type Record struct {
	StatusCode int //if in future need to cache other status codes
	Body       []byte
	Headers    http.Header
	expiry     time.Time
	size       int

	linkedlistEle *list.Element
}

func NewMemoryCache(ttl time.Duration, maxSize, maxRecordSize int) *MemoryCache {
	return &MemoryCache{
		records:           make(map[string]*Record),
		ttl:               ttl,
		maxCacheSize:      maxSize,
		maxRecordSize:     maxRecordSize,
		remainingCapacity: maxSize,
		ll:                list.New(),
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
		cache.ll.Remove(r.linkedlistEle)
		delete(cache.records, k)

		return nil
	}

	cache.ll.MoveToFront(r.linkedlistEle)

	return r
}

func (cache *MemoryCache) Set(k string, data *Record) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	data.size = data.Calsize()

	//not caching request if it exceeds maxRequestSize limit
	if data.size > cache.maxRecordSize {
		return ErrMaxRecordSizeExceed
	}

	existingRecord, ok := cache.records[k]
	if ok {
		oldSize := existingRecord.size
		cache.remainingCapacity += oldSize
	}

	//keep deleting old entry in cache as cache has reached it's limit
	for data.size > cache.remainingCapacity {
		oldest := cache.ll.Back()
		oldestRecordKey := oldest.Value.(string)

		oldestRecord := cache.records[oldestRecordKey]
		cache.remainingCapacity += oldestRecord.size
		delete(cache.records, oldestRecordKey)

		cache.ll.Remove(oldest)
	}

	cache.remainingCapacity -= data.size

	data.expiry = time.Now().Add(cache.ttl)

	data.linkedlistEle = cache.ll.PushFront(k)
	cache.records[k] = data

	return nil
}

func (cache *MemoryCache) Count() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return len(cache.records)
}

func (r *Record) Calsize() int {
	size := 40 // accounting for 2 int variables and one time.Time field

	for k, vals := range r.Headers {
		size += len(k)
		for _, v := range vals {
			size += len(v)
		}
	}
	size += len(r.Body)

	return size
}
