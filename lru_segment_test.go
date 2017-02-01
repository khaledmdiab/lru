package lru_test

import (
	"testing"

	"github.com/khaledmdiab/lru"
)

func TestSegmentCache(t *testing.T) {
	var evicted []string
	lru := lru.NewSegmentCache(3, func(key string) {
		evicted = append(evicted, key)
	})
	lru.Add("1", "one", 1)
	lru.Add("2", "two", 1)
	lru.Add("3", "three", 1)

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
	lru.Add("4", "four",2)

	// LRU here is 4, 1
	_, ok = lru.Get("2")
	if ok {
		t.Error("2: in cache")
	}

	if len(evicted) != 2 || evicted[0] != "2" || evicted[1] != "3" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// Add new item.
	lru.Add("5", "five", 1)

	// LRU here is 5, 4
	ok = lru.HasKey("5")
	if !ok {
		t.Error("5: not in cache")
	}

	if len(evicted) != 1 || evicted[0] != "1" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// LRU is now 1, 5
	lru.Add("1", "one", 1)

	// LRU is now 2, 1, 5
	lru.Add("2", "two", 1)

	if len(evicted) != 1 || evicted[0] != "4" {
		t.Error("evicted: incorrect")
	}
	evicted = nil

	// LRU is now 4, 2
	lru.Add("4", "four", 2)
	if lru.Len() != 2 {
		t.Error("len != 2")
	}

	// LRU is now 5, 4
	lru.Add("5", "five", 1)
	if lru.Len() != 2 {
		t.Error("len != 2")
	}

	// The LRU queue should be 5, 4
	if len(evicted) != 3 || evicted[0] != "5" || evicted[1] != "1" || evicted[2] != "2" {
		t.Error("evicted: incorrect")
	}
	evicted = nil
	t.Log("okay")
}

func TestSegmentCacheThreadSafe(t *testing.T) {
	//t.Log("This does NOT test (un)pinning items...")

}