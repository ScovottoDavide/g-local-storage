package glocalstorage

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type backgroundCleaner struct {
	Interval time.Duration
	stop     chan bool
}

type node struct {
	key        string
	value      []byte
	expiration *time.Time
	next       *node
	prev       *node
}

type CacheItem struct {
	value      []byte
	expiration *time.Time
}

type StorageConfig struct {
	Expiration      time.Duration // if set = 0 items will never expire
	Capacity        int64
	CleanupInterval time.Duration // if set > 0 will run a background storage cleaner
}

type InternalLocalStorage struct {
	kv_storage     map[string]*node
	head           *node
	tail           *node
	size           int64
	defaultConfigs StorageConfig
	lock           *sync.Mutex
	cleaner        *backgroundCleaner
}

type LocalStorage struct {
	*InternalLocalStorage
}

func New(config StorageConfig) *LocalStorage {
	intLocalStorage := &InternalLocalStorage{
		kv_storage:     make(map[string]*node),
		head:           nil,
		tail:           nil,
		size:           0,
		defaultConfigs: config,
		lock:           &sync.Mutex{},
	}
	localStorage := &LocalStorage{InternalLocalStorage: intLocalStorage}

	cleanupInterval := config.CleanupInterval

	if cleanupInterval > 0 && config.Expiration > 0 { // run background cleaner iff nodes can actually expire
		runCleaner(intLocalStorage, cleanupInterval)
		runtime.SetFinalizer(localStorage, stopCleaner)
	} else {
		fmt.Println("Background cleaner has NOT been SET")
	}
	return localStorage
}

func runCleaner(intLocalStorage *InternalLocalStorage, cleanupInterval time.Duration) {
	cleaner := &backgroundCleaner{
		Interval: cleanupInterval,
		stop:     make(chan bool),
	}

	intLocalStorage.cleaner = cleaner
	go cleaner.Run(intLocalStorage)
}

func (cleaner *backgroundCleaner) Run(intLocalStorage *InternalLocalStorage) {
	ticker := time.NewTicker(cleaner.Interval)
	for {
		select {
		case <-ticker.C:
			fmt.Println("STARTING Backgroung cleaner")
			intLocalStorage.CleanUpExpired()
		case <-cleaner.stop:
			ticker.Stop()
			fmt.Println("STOPPED Backgroung cleaner")
			return
		}
	}
}

func stopCleaner(c *LocalStorage) {
	fmt.Println("Stopping backgroung cleaner")
	c.cleaner.stop <- true
}

func (local_storage *InternalLocalStorage) Set(key string, value []byte, itemExpiration time.Duration) (updated bool) {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	var expiration *time.Time
	if itemExpiration > 0 {
		tmp := time.Now().Add(itemExpiration)
		expiration = &tmp
	} else if itemExpiration == 0 {
		expiration = nil
	} else if local_storage.defaultConfigs.Expiration > 0 && itemExpiration == -1 {
		tmp := time.Now().Add(local_storage.defaultConfigs.Expiration)
		expiration = &tmp
	} else {
		expiration = nil
	}

	// key does not exist --> insert
	if local_storage.kv_storage[key] == nil {
		if local_storage.head == nil && local_storage.tail == nil { // first element
			tmp_node := &node{
				key:        key,
				value:      value,
				expiration: expiration, // results in the time when the value expires: if defaultExpiration = 1day and now is 24Jan --> expiration = 25Jan
				next:       nil,
				prev:       nil,
			}
			local_storage.head = tmp_node
			local_storage.tail = local_storage.head
			local_storage.kv_storage[key] = tmp_node
		} else {
			if local_storage.size == local_storage.defaultConfigs.Capacity {
				local_storage.evict()
			}
			tmp_node := &node{
				key:        key,
				value:      value,
				expiration: expiration, // results in the time when the value expires: if defaultExpiration = 1day and now is 24Jan --> expiration = 25Jan
				next:       local_storage.head,
				prev:       nil,
			}
			local_storage.head.prev = tmp_node
			local_storage.head = tmp_node
			local_storage.kv_storage[key] = tmp_node
		}
		local_storage.size++
		return false
	} else { // key does already exist --> update value and bring up the updated node to be the new head
		// update value
		newHead := local_storage.kv_storage[key]
		newHead.value = value

		// update head
		local_storage.updateHead(newHead)
		return true
	}
}

func (local_storage *InternalLocalStorage) Get(key string) (cacheItem *CacheItem, hit bool) {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	newHead := local_storage.kv_storage[key]

	if newHead == nil {
		return nil, false
	} else if local_storage.isNodeExpired(newHead) {
		local_storage.removeNode(newHead)
		return nil, false
	}

	local_storage.updateHead(newHead)
	return &CacheItem{value: newHead.value, expiration: newHead.expiration}, true
}

func (local_storage *InternalLocalStorage) Delete(key string) bool {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	toRemove := local_storage.kv_storage[key]

	if toRemove == nil {
		return false
	}
	local_storage.removeNode(toRemove)
	return true
}

func (local_storage *InternalLocalStorage) Clear() {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	for i := local_storage.tail; i != nil; {
		local_storage.removeNode(i)
		i = local_storage.tail
	}
}

func (local_storage *InternalLocalStorage) Show() {
	for i := local_storage.head; i != nil; {
		fmt.Println("Key: ", i.key)
		fmt.Println("Value: ", string(i.value))
		fmt.Println("ttl: ", i.expiration)
		i = i.next
	}
}

// LRU policy eviction
func (local_storage *InternalLocalStorage) evict() {
	to_evict := local_storage.tail
	local_storage.tail = to_evict.prev
	local_storage.tail.next = nil
	local_storage.size--

	to_evict.next = nil
	to_evict.prev = nil
	delete(local_storage.kv_storage, to_evict.key)
}

func (local_storage *InternalLocalStorage) updateHead(newHead *node) {
	if local_storage.head == local_storage.tail {
		return
	}
	prev_ := newHead.prev
	prev_.next = newHead.next
	if prev_.next == nil {
		local_storage.tail = prev_
	}
	newHead.next = local_storage.head
	newHead.prev = nil
	local_storage.head = newHead
	local_storage.head.next.prev = local_storage.head
}

func (local_storage *InternalLocalStorage) CleanUpExpired() {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()
	for i := local_storage.head; i != nil; {
		if local_storage.isNodeExpired(i) {
			local_storage.removeNode(i)
		}
		i = i.next
	}
}

func (local_storage *InternalLocalStorage) isNodeExpired(node *node) bool {
	if node.expiration != nil {
		return node.expiration.Before(time.Now())
	}
	return false
}

func (local_storage *InternalLocalStorage) removeNode(nodeToRemove *node) {
	if nodeToRemove.prev == nil && nodeToRemove.next == nil { // last elem
		local_storage.head = nil
		local_storage.tail = nil
	} else if nodeToRemove.prev == nil { // head
		local_storage.head = nodeToRemove.next
		local_storage.head.prev = nil
	} else if nodeToRemove.next == nil { // tail
		local_storage.tail = nodeToRemove.prev
		local_storage.tail.next = nil
	} else {
		nodeToRemove.prev.next = nodeToRemove.next
		nodeToRemove.next.prev = nodeToRemove.prev
	}

	nodeToRemove.next = nil
	nodeToRemove.prev = nil
	delete(local_storage.kv_storage, nodeToRemove.key)
	local_storage.size--
}
