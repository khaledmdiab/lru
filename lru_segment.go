package lru

import (
	"container/list"
	"fmt"
	"sync"
)

// The id cache entry element. Each element is a video segment
type cacheSegmentEntry struct {
	// LRU Entry key and value.
	key string

	// Segment size
	size int64

	// The associated data.
	data interface{}

	// Position in the LRU queue. If the entry is pinned this is nil.
	position *list.Element
}

// LRUCache is a least recently used cache implementation with pinned
// members.  Pinned members do not count in the size of the cache when
// deciding when to evict cache entries.
type SegmentCache struct {
	// Available capacity of LRU cache.
	capacity int64

	// Used capacity so far
	usedCapacity int64

	// SegmentCache of entries for O(1) lookup.
	cache map[string]*cacheSegmentEntry

	// Queue.
	q *list.List

	// Callback for eviction.
	evictedCallback func(key string)

	// Read/Write mutex
	lock sync.RWMutex
}

// NewLRUCache creates a new cache of the given size.
func NewSegmentCache(capacity int64, evictedCallback func(key string)) *SegmentCache {
	lru := &SegmentCache{
		capacity:            capacity,
		usedCapacity:        0,
		cache:               make(map[string]*cacheSegmentEntry),
		q:                   list.New(),
		evictedCallback:     evictedCallback}
	return lru

}

// Len returns the number of segments in the cache.
func (lru *SegmentCache) Len() int {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	return len(lru.cache)
}

// Capacity returns the capacity of the cache.
func (lru *SegmentCache) Capacity() int64 {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	return lru.capacity
}

// UsedCapacity returns the used capacity of the cache.
func (lru *SegmentCache) UsedCapacity() int64 {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	return lru.usedCapacity
}

// Get an item from the cache. Moves the item to the front of the queue
// if not pinned. Returns (item, true) if in the cache, (nil, false) otherwise.
func (lru *SegmentCache) Get(key string) (interface{}, bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if e, ok := lru.cache[key]; ok {
		// Move to the front of the list.
		lru.q.MoveToFront(e.position)
		return e.data, true
	}
	return nil, false
}

// HasKey determines whether the given key is in the cache without changing LRU order.
func (lru *SegmentCache) HasKey(key string) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	_, ok := lru.cache[key]
	return ok
}

// Add a new item to the queue, evicting an item from the cache
// if full.
func (lru *SegmentCache) Add(key string, data interface{}, size int64) {
	lru.lock.Lock()
	// Check for existing item, replacing the data if already
	// present.
	if e, ok := lru.cache[key]; ok {
		lru.q.MoveToFront(e.position)
		e.data = data
		lru.lock.Unlock()
		return
	}

	entry := &cacheSegmentEntry{key: key, data: data, size: size}
	entry.position = lru.q.PushFront(entry)
	lru.usedCapacity += size
	lru.cache[key] = entry
	lru.lock.Unlock()
	lru.evict()
}

// PrintStats prints information on the cache.
func (lru *SegmentCache) PrintStats() {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	fmt.Printf("LRU used capacity: %d\n", lru.usedCapacity)
}

// Evict the least recently used item from the cache.
func (lru *SegmentCache) evict() {
	lru.lock.Lock()
	for lru.usedCapacity > lru.capacity {
		e := lru.q.Remove(lru.q.Back()).(*cacheSegmentEntry)
		lru.usedCapacity -= e.size
		delete(lru.cache, e.key)
		if lru.evictedCallback != nil {
			lru.evictedCallback(e.key)
		}
	}
	lru.lock.Unlock()

	lru.PrintStats()
}
