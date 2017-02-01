package lru_test

import (
	"testing"

	"github.com/khaledmdiab/lru"
)

func TestCache(t *testing.T) {
	var evicted []string
	lru := lru.NewCache(3, func(key string) {
		evicted = append(evicted, key)
	})
	lru.Add("1", "one")
	lru.Add("2", "two")
	lru.Add("3", "three")
	// LRU here is 3, 2, 1
	if lru.Len() != 3 {
		t.Error("len != 3")
	}

	_, ok := lru.Get("0")
	if ok {
		t.Error("0: in cache")
	}
	val, ok := lru.Get("1")
	if !ok {
		t.Error("1: not in cache")
	}
	if val != "one" {
		t.Error("val: not \"one\"")
	}
	if len(evicted) != 0 {
		t.Error("evicted: not empty")
	}

	// LRU here is 1, 3, 2
	lru.Add("4", "four")

	// LRU here is 4, 1, 3
	_, ok = lru.Get("2")
	if ok {
		t.Error("2: in cache")
	}

	if len(evicted) != 1 || evicted[0] != "2" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// Pin 3.
	lru.Pin("3")
	// LRU here is 4, 1 | 3
	_, ok = lru.Get("3")
	if !ok {
		t.Error("3: not in cache")
	}

	// Add new item.
	lru.Add("5", "five")
	// LRU here is 5, 4, 1 | 3
	ok = lru.HasKey("1")
	if !ok {
		t.Error("1: not in cache")
	}

	if len(evicted) != 0 {
		t.Error("evicted: not empty")
	}
	evicted = nil

	lru.Unpin("3")
	// LRU here is 3, 5, 4
	if len(evicted) != 1 || evicted[0] != "1" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// LRU is now 3, 5, 4
	lru.Add("1", "one")
	// LRU is now 1, 3, 5
	lru.Add("2", "two")
	// LRU is now 2, 1, 3

	if len(evicted) != 2 || evicted[0] != "4" || evicted[1] != "5" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// Pin all the items in the cache.
	lru.Pin("2")
	lru.Pin("1")
	lru.Pin("3")

	lru.Add("4", "four")
	if lru.Len() != 4 {
		t.Error("len != 4")
	}
	lru.Pin("4")
	lru.Add("5", "five")
	if lru.Len() != 5 {
		t.Error("len != 5")
	}
	// LRU is now 5 | 2 1 3 4

	// Unpin the items.
	lru.Unpin("1")
	lru.Unpin("2")
	lru.Unpin("3")
	lru.Unpin("4")

	// The LRU queue should be 4, 3, 2
	if len(evicted) != 2 || evicted[0] != "5" || evicted[1] != "1" {
		t.Error("evicted: incorrect")
	}
	evicted = nil
}

func TestCacheThreadSafe(t *testing.T) {
	t.Log("This does NOT test (un)pinning items...")

}