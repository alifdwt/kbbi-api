package cache

import (
	"sync"
	"time"
)

var mu sync.RWMutex
var store = map[string][]byte{}
var expiry = map[string]time.Time{}

func Set(key string, val []byte, ttl time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	store[key] = val
	if ttl > 0 {
		expiry[key] = time.Now().Add(ttl)
	} else {
		delete(expiry, key)
	}
}

func Get(key string) ([]byte, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if t, ok := expiry[key]; ok {
		if time.Now().After(t) {
			return nil, false
		}
	}

	v, ok := store[key]
	return v, ok
}
