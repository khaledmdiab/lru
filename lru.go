package lru

import (
	"container/list"
	"fmt"
)

// The id cache entry element.
type cacheEntry struct {
	// LRU Entry key and value.
	key string

	// The associated data.
	data interface{}

	// If true the entry cannot be evicted.
	pinned bool

	// Position in the LRU queue. If the entry is pinned this is nil.
	position *list.Element
}

// LRUCache is a least recently used cache implementation with pinned
// members.  Pinned members do not count in the size of the cache when
// deciding when to evict cache entries.
type Cache struct {
	// Number of entries in the LRU cache.
	size int

	// Cache of entries for O(1) lookup.
	cache map[string]*cacheEntry

	// Queue.
	q *list.List

	// Callback for eviction.
	evictedCallback func(key string)
}

// NewLRUCache creates a new cache of the given size.
func NewCache(size int, evictedCallback func(key string)) *Cache {
	lru := &Cache{
		size:            size,
		cache:           make(map[string]*cacheEntry),
		q:               list.New(),
		evictedCallback: evictedCallback}
	return lru

}

// Len returns the number of items in the cache. This can be greater than the
// size due to pinned items.
func (lru *Cache) Len() int {
	return len(lru.cache)
}

// Size returns the size of the cache.
func (lru *Cache) Size() int {
	return lru.size
}

// Get an item from the cache. Moves the item to the front of the queue
// if not pinned. Returns (item, true) if in the cache, (nil, false) otherwise.
func (lru *Cache) Get(key string) (interface{}, bool) {
	if e, ok := lru.cache[key]; ok {
		// If the item isn't pinned move to the front of the list.
		if !e.pinned {
			lru.q.MoveToFront(e.position)
		}

		return e.data, true
	}
	return nil, false
}

// HasKey determines whether the given key is in the cache without changing LRU order.
func (lru *Cache) HasKey(key string) bool {
	_, ok := lru.cache[key]
	return ok
}

// Add a new item to the queue, evicting an item from the cache
// if full.
func (lru *Cache) Add(key string, data interface{}) {
	// Check for existing item, replacing the data if already
	// present.
	if e, ok := lru.cache[key]; ok {
		if !e.pinned {
			lru.q.MoveToFront(e.position)
		}
		e.data = data
		return
	}

	entry := &cacheEntry{key: key, data: data}
	entry.position = lru.q.PushFront(entry)

	lru.cache[key] = entry

	lru.evict()
}

// Pin ensures that the item with the given key is not evicted from
// the cache. Pinned items do not count torwards the cache size.
func (lru *Cache) Pin(key string) {
	if e, ok := lru.cache[key]; ok {
		if !e.pinned {
			e.pinned = true
			lru.q.Remove(e.position)
			e.position = nil
		}
	}
}

// Unpin removes the cache pin from the item with the given key.
// The unpinned item is placed at the head of the cache.
func (lru *Cache) Unpin(key string) {
	if e, ok := lru.cache[key]; ok {
		if e.pinned {
			e.pinned = false
			e.position = lru.q.PushFront(e)
			lru.evict()
		}
	}
}

// IsPinned returns true if the key is pinned, false otherwise.
func (lru *Cache) IsPinned(key string) (bool, error) {
	if e, ok := lru.cache[key]; ok {
		return e.pinned, nil
	}
	return false, fmt.Errorf("%s: not in cache", key)
}

// PrintStats prints information on the cache.
func (lru *Cache) PrintStats() {
	fmt.Printf("%d records, %d in queue, %d pinned\n", len(lru.cache), lru.q.Len(), len(lru.cache)-lru.q.Len())
}

// Evict the least recently used item from the cache.
func (lru *Cache) evict() {
	if lru.q.Len() > lru.size {
		e := lru.q.Remove(lru.q.Back()).(*cacheEntry)
		delete(lru.cache, e.key)
		if lru.evictedCallback != nil {
			lru.evictedCallback(e.key)
		}
	}
}
