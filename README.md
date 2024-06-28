# Local Cache Implementation
This repository contains an implementation of a local cache using a hash map and a doubly linked list. The cache utilizes time-to-live (TTL) to delete expired values and implements a Least-Recently-Used (LRU) policy to evict values in case the cache is full.

### Features
- StorageConfig Struct: Defines configuration parameters such as expiration time and capacity for the cache.
- LocalStorage Struct: Represents the local cache with methods for setting, getting, deleting, clearing, and displaying cache contents.
- Node Struct: Represents a node in the cache with key, value, expiration time, and pointers to the next and previous nodes.
- New Function: Initializes a new local cache with the given configuration.
- Set Method: Inserts a key-value pair into the cache, updating the expiration time and handling eviction if necessary.
- Get Method: Retrieves a value from the cache based on the given key, updating its access time.
- Delete Method: Removes a key-value pair from the cache.
- Clear Method: Clears all entries in the cache.
- ShowCache Method: Displays the contents of the cache.

### Usage
To use this local cache implementation, follow these steps:

1. Import the glocalstorage package.
2. Initialize a new cache with the desired configuration using the New function.
Use the Set, Get, Delete, Clear, and ShowCache methods to interact with the cache.
go

```golang
import (
    "fmt"
    "github.com/ScovottoDavide/glocalstorage"
    "time"
)

func main() {
    config := glocalstorage.StorageConfig{
        Expiration: 24 * time.Hour,
        Capacity:   100,
    }

    cache := glocalstorage.New(config)

    cache.Set("key", []byte("value"))
    value, _ := cache.Get("key")
    fmt.Println("Retrieved value:", string(value))
}
```
