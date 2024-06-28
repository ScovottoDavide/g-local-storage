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

	cache.Set("key1", []byte("1abcdefg"))
	cache.Set("key2", []byte("2abcdefg"))
	cache.Set("key3", []byte("3abcdefg"))
	cache.Set("key4", []byte("4abcdefg"))

	curSize := cache.size
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}

	cache.Set("key1", []byte("1abcdefghilm"))
	cache.ShowCache()
	if cache.head.key != "key1" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 4 {
		t.Errorf("curSize should be 4, instead got %d\n", curSize)
	}

	cache.Set("key4", []byte("4abcdefghilm"))
	cache.ShowCache()
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

	cache.Set("key1", []byte("1abcdefg"))
	cache.Set("key2", []byte("2abcdefg"))
	cache.Set("key3", []byte("3abcdefg"))
	cache.Set("key4", []byte("4abcdefg"))
	cache.Set("key5", []byte("5abcdefg"))

	curSize := cache.size
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}

	cache.Set("key6", []byte("6abcdefg"))
	if cache.head.key != "key6" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}

	cache.Set("key7", []byte("7abcdefg"))
	if cache.head.key != "key7" {
		t.Errorf("Head key is supposed to be key1, instead got %s\n", cache.head.key)
	}
	if curSize != 5 {
		t.Errorf("curSize should be 5, instead got %d\n", curSize)
	}

	cache.ShowCache()
}

func TestCacheDelete(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"))
	cache.Set("key2", []byte("2abcdefg"))
	cache.Set("key3", []byte("3abcdefg"))
	cache.Set("key4", []byte("4abcdefg"))
	cache.Set("key5", []byte("5abcdefg"))

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

	cache.ShowCache()
}

func TestCacheClear(t *testing.T) {
	storageExpiration := time.Duration(time.Hour * 24)
	var capacity int64 = 5

	config := StorageConfig{Expiration: storageExpiration, Capacity: capacity}
	cache := New(config)

	cache.Set("key1", []byte("1abcdefg"))
	cache.Set("key2", []byte("2abcdefg"))
	cache.Set("key3", []byte("3abcdefg"))
	cache.Set("key4", []byte("4abcdefg"))
	cache.Set("key5", []byte("5abcdefg"))

	cache.Clear()
	curSize := cache.size
	if curSize != 0 {
		t.Errorf("curSize should be 0, instead got %d\n", curSize)
	}
	cache.ShowCache()
}
