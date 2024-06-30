package glocalstorage

import (
	"testing"
	"time"
)

func TestLocalStorageSet(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 10

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"), -1)
	cache.Set("key2", []byte("2abcdefg"), -1)
	cache.Set("key3", []byte("3abcdefg"), -1)
	cache.Set("key4", []byte("4abcdefg"), -1)

	curSize := cache.size
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}

	cache.Set("key1", []byte("1abcdefghilm"), -1)
	if cache.head.key != "key1" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}

	cache.Set("key4", []byte("4abcdefghilm"), -1)
	if cache.head.key != "key4" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}
}

func TestLocalStorageLRU(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"), -1)
	cache.Set("key2", []byte("2abcdefg"), -1)
	cache.Set("key3", []byte("3abcdefg"), -1)
	cache.Set("key4", []byte("4abcdefg"), -1)
	cache.Set("key5", []byte("5abcdefg"), -1)

	curSize := cache.size
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}

	cache.Set("key6", []byte("6abcdefg"), -1)
	if cache.head.key != "key6" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}

	cache.Set("key7", []byte("7abcdefg"), -1)
	if cache.head.key != "key7" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}
}

func TestCacheDelete(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"), -1)
	cache.Set("key2", []byte("2abcdefg"), -1)
	cache.Set("key3", []byte("3abcdefg"), -1)
	cache.Set("key4", []byte("4abcdefg"), -1)
	cache.Set("key5", []byte("5abcdefg"), -1)

	cache.Delete("key1")
	curSize := cache.size
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}

	cache.Delete("key3")
	curSize = cache.size
	if curSize != 3 {
		t.Errorf("curSize should be 3, instead got %d\n", curSize)
	}
}

func TestCacheClear(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"), -1)
	cache.Set("key2", []byte("2abcdefg"), -1)
	cache.Set("key3", []byte("3abcdefg"), -1)
	cache.Set("key4", []byte("4abcdefg"), -1)
	cache.Set("key5", []byte("5abcdefg"), -1)

	cache.Clear()
	curSize := cache.size
	if curSize != 0 {
		t.Errorf("curSize should be 0, instead got %d\n", curSize)
	}
}

func TestCacheExpiration(t *testing.T) {
	// set a low cache expiration time
	storageExpiration := time.Duration(time.Second * 2)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"), -1)

	// let cache value expire
	t.Log("Sleeping for 5 seconds...")
	time.Sleep(time.Second * 5)

	// Try to GET key from cache. It should remove the expired key
	cacheItem, hit := cache.Get("key1")
	if cacheItem != nil && hit != false {
		t.Errorf("Cache did not evict key key1.")
	}
}

func TestCacheWithCleaner(t *testing.T) {
	storageExpiration := time.Duration(time.Second * 2)
	cleanupInterval := time.Duration(time.Second * 1)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity, CleanupInterval: cleanupInterval}
	cache := New(config)

	cache.Set("key1", []byte("ABCDEFG"), -1)

	<-time.After(time.Second * 3)

	cacheItem, hit := cache.Get("key1")
	if cacheItem != nil && hit != false {
		t.Errorf("Background cleaner DID NOT remove expired nodes.")
	}
}

func TestCacheWithSlowCleaner(t *testing.T) {
	storageExpiration := time.Duration(time.Second * 2)
	cleanupInterval := time.Duration(time.Second * 5)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity, CleanupInterval: cleanupInterval}
	cache := New(config)

	cache.Set("key1", []byte("ABCDEFG"), -1)

	<-time.After(time.Second * 3)

	// Also when the cleanup Interval is higher than the storage expiration, the Get function always checks if the value is not expired.
	cacheItem, hit := cache.Get("key1")
	if cacheItem != nil && hit != false {
		t.Errorf("Background cleaner DID NOT remove expired nodes.")
	}
}

func TestCacheWithNoExpiration(t *testing.T) {
	cleanupInterval := time.Duration(time.Second * 5)
	var capacity int64 = 5

	config := StorageConfig{Expiration: 0, Capacity: capacity, CleanupInterval: cleanupInterval}
	cache := New(config)

	cache.Set("key1", []byte("ABCDEFG"), -1)
	cacheItem, hit := cache.Get("key1")
	if cacheItem == nil && hit == false {
		t.Errorf("Node with key1 is not in cache. ERROR")
	} else {
		if cacheItem.Expiration != nil {
			t.Errorf("Node should NEVER EXPIRE. ERROR")
		}
	}

	cache.CleanUpExpired()
	if cache.size != 1 {
		t.Errorf("ERROR. Removed non expired key1")
	}
}
