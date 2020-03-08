package fzflib

import "sync"

// queryCache associates strings to lists of items
type queryCache map[string][]result

// chunkCache associates chunk and query string to lists of items
type chunkCache struct {
	mutex sync.Mutex
	cache map[*chunk]*queryCache
}

// newChunkCache returns a new chunkCache
func newChunkCache() chunkCache {
	return chunkCache{sync.Mutex{}, make(map[*chunk]*queryCache)}
}

// Add adds the list to the cache
func (cc *chunkCache) Add(chunk *chunk, key string, list []result) {
	if len(key) == 0 || !chunk.IsFull() || len(list) > queryCacheMax {
		return
	}

	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	qc, ok := cc.cache[chunk]
	if !ok {
		cc.cache[chunk] = &queryCache{}
		qc = cc.cache[chunk]
	}
	(*qc)[key] = list
}

// Lookup is called to lookup chunkCache
func (cc *chunkCache) Lookup(chunk *chunk, key string) []result {
	if len(key) == 0 || !chunk.IsFull() {
		return nil
	}

	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	qc, ok := cc.cache[chunk]
	if ok {
		list, ok := (*qc)[key]
		if ok {
			return list
		}
	}
	return nil
}

func (cc *chunkCache) Search(chunk *chunk, key string) []result {
	if len(key) == 0 || !chunk.IsFull() {
		return nil
	}

	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	qc, ok := cc.cache[chunk]
	if !ok {
		return nil
	}

	for idx := 1; idx < len(key); idx++ {
		// [---------| ] | [ |---------]
		// [--------|  ] | [  |--------]
		// [-------|   ] | [   |-------]
		prefix := key[:len(key)-idx]
		suffix := key[idx:]
		for _, substr := range [2]string{prefix, suffix} {
			if cached, found := (*qc)[substr]; found {
				return cached
			}
		}
	}
	return nil
}
