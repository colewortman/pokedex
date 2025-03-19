package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt	time.Time
	val			[]byte
}

type Cache struct {
	entries map[string]cacheEntry
	mutex 	sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache {
		entries: make(map[string]cacheEntry),
	}

	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := cacheEntry{
		createdAt:  time.Now(),
		val:		val,
	}

	c.entries[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, found := c.entries[key]
	if !found {
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		<-ticker.C
		c.reap(interval)
	}
}

func (c *Cache) reap(interval time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) > interval {
			delete(c.entries, key)
		}
	}
}