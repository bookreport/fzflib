package fzflib

import "sync"

// chunk is a list of Items whose size has the upper limit of chunkSize
type chunk struct {
	items [chunkSize]item
	count int
}

// itemBuilder is a closure type that builds item object from byte array
type itemBuilder func(*item, []byte) bool

// chunkList is a list of Chunks
type chunkList struct {
	chunks []*chunk
	mutex  sync.Mutex
	trans  itemBuilder
}

// newChunkList returns a new chunkList
func newChunkList(trans itemBuilder) *chunkList {
	return &chunkList{
		chunks: []*chunk{},
		mutex:  sync.Mutex{},
		trans:  trans}
}

func (c *chunk) push(trans itemBuilder, data []byte) bool {
	if trans(&c.items[c.count], data) {
		c.count++
		return true
	}
	return false
}

// IsFull returns true if the chunk is full
func (c *chunk) IsFull() bool {
	return c.count == chunkSize
}

func (cl *chunkList) lastChunk() *chunk {
	return cl.chunks[len(cl.chunks)-1]
}

// countItems returns the total number of Items
func countItems(cs []*chunk) int {
	if len(cs) == 0 {
		return 0
	}
	return chunkSize*(len(cs)-1) + cs[len(cs)-1].count
}

// Push adds the item to the list
func (cl *chunkList) Push(data []byte) bool {
	cl.mutex.Lock()

	if len(cl.chunks) == 0 || cl.lastChunk().IsFull() {
		cl.chunks = append(cl.chunks, &chunk{})
	}

	ret := cl.lastChunk().push(cl.trans, data)
	cl.mutex.Unlock()
	return ret
}

// Clear clears the data
func (cl *chunkList) Clear() {
	cl.mutex.Lock()
	cl.chunks = nil
	cl.mutex.Unlock()
}

// Snapshot returns immutable snapshot of the chunkList
func (cl *chunkList) Snapshot() ([]*chunk, int) {
	cl.mutex.Lock()

	ret := make([]*chunk, len(cl.chunks))
	copy(ret, cl.chunks)

	// Duplicate the last chunk
	if cnt := len(ret); cnt > 0 {
		newChunk := *ret[cnt-1]
		ret[cnt-1] = &newChunk
	}

	cl.mutex.Unlock()
	return ret, countItems(ret)
}
