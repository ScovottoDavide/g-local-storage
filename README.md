# Local Cache Implementation
This repository contains an implementation of a local cache using a hash map and a doubly linked list. The cache utilizes time-to-live (TTL) to delete expired values and implements a Least-Recently-Used (LRU) policy to evict values in case the cache is full.

Additionally, a background cleaner has been implemented to periodically clear out expired values from the cache. This can be enabled by properly configuring the `LocalStorage`:

```go
config := glocalstorage.StorageConfig{
    Expiration: 24 * time.Hour,
    Capacity:   100,
    CleanupInterval: 1 * time.Hour, // Cleanup runs in the background every hour
}
```

If the `CleanupInterval` is set to `0`, the `BackgroundCleaner` goroutine will not be fired. Hence, this can in a higher memory usage becuase items will not be deleted until they are queried (using `cache.Get`) or until the LRU policy decides to `evict` an expired item.


### Features
- StorageConfig Struct: Defines configuration parameters such as expiration time and capacity for the cache.
- LocalStorage Struct: Represents the local cache with methods for setting, getting, deleting, clearing, and displaying cache contents.
- Node Struct: Represents a node in the cache with key, value, expiration time, and pointers to the next and previous nodes.
- New Function: Initializes a new local cache with the given configuration.
- Set Method: Inserts a key-value pair into the cache, updating the expiration time and handling eviction if necessary.
- Get Method: Retrieves a value from the cache based on the given key, updating its access time.
- Delete Method: Removes a key-value pair from the cache.
- Clear Method: Clears all entries in the cache.
- Show Method: Displays the contents of the cache.

### Usage
To use this local cache implementation, follow these steps:

1. Import the g-local-storage package.
2. Initialize a new cache with the desired configuration using the New function.
Use the Set, Get, Delete, Clear, and Show methods to interact with the cache.

```golang
import (
    "fmt"
    "github.com/ScovottoDavide/g-local-storage"
    "time"
)

func main() {
    config := glocalstorage.StorageConfig{
        Expiration: 24 * time.Hour,
        Capacity:   100,
        CleanupInterval: 1 * time.Hour, // Cleanup runs in the background every hour
    }

    cache := glocalstorage.New(config)

    cache.Set("key", []byte("value"))
    value, _ := cache.Get("key")
    fmt.Println("Retrieved value:", string(value))
}
```
