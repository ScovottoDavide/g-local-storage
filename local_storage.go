package glocalstorage

import (
	"fmt"
	"sync"
	"time"
)

type node struct {
	key        string
	value      []byte
	expiration time.Time
	next       *node
	prev       *node
}

type StorageConfig struct {
	Expiration time.Duration
	Capacity   int64
}

type LocalStorage struct {
	kv_storage     map[string]*node
	head           *node
	tail           *node
	size           int64
	defaultConfigs StorageConfig
	lock           *sync.Mutex
}

func New(config StorageConfig) *LocalStorage {
	return &LocalStorage{
		kv_storage:     make(map[string]*node),
		head:           nil,
		tail:           nil,
		size:           0,
		defaultConfigs: config,
		lock:           &sync.Mutex{},
	}
}

func (local_storage *LocalStorage) Set(key string, value []byte) (updated bool) {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	// key does not exist --> insert
	if local_storage.kv_storage[key] == nil {
		if local_storage.head == nil && local_storage.tail == nil { // first element
			tmp_node := &node{
				key:        key,
				value:      value,
				expiration: time.Now().Add(local_storage.defaultConfigs.Expiration), // results in the time when the value expires: if defaultExpiration = 1day and now is 24Jan --> expiration = 25Jan
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
				expiration: time.Now().Add(local_storage.defaultConfigs.Expiration), // results in the time when the value expires: if defaultExpiration = 1day and now is 24Jan --> expiration = 25Jan
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

func (local_storage *LocalStorage) Get(key string) (value []byte, hit bool) {
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
	return newHead.value, true
}

func (local_storage *LocalStorage) Delete(key string) bool {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	toRemove := local_storage.kv_storage[key]

	if toRemove == nil {
		return false
	}
	local_storage.removeNode(toRemove)
	return true
}

func (local_storage *LocalStorage) Clear() {
	local_storage.lock.Lock()
	defer local_storage.lock.Unlock()

	for i := local_storage.tail; i != nil; {
		local_storage.removeNode(i)
		i = local_storage.tail
	}
}

func (local_storage *LocalStorage) Show() {
	for i := local_storage.head; i != nil; {
		fmt.Println("Key: ", i.key)
		fmt.Println("Value: ", string(i.value))
		fmt.Println("ttl: ", i.expiration)
		i = i.next
	}
}

// LRU policy eviction
func (local_storage *LocalStorage) evict() {
	to_evict := local_storage.tail
	local_storage.tail = to_evict.prev
	local_storage.tail.next = nil
	local_storage.size--

	to_evict.next = nil
	to_evict.prev = nil
	delete(local_storage.kv_storage, to_evict.key)
}

func (local_storage *LocalStorage) updateHead(newHead *node) {
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

func (local_storage *LocalStorage) isNodeExpired(node *node) bool {
	return node.expiration.Before(time.Now())
}

func (local_storage *LocalStorage) removeNode(nodeToRemove *node) {
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
