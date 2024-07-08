package docker

import (
	"sync"
	"time"
)

type ImageQueue struct {
	// array of items
	items []ImageInfo
	// Mutual exclusion lock
	lock sync.Mutex
	// Cond is used to pause mulitple goroutines and wait
	cond *sync.Cond
	// Total Size of the images in the queue
	totalSize int64
}

// Initialize ConcurrentQueue
func NewImageQueue() *ImageQueue {
	q := &ImageQueue{}
	q.cond = sync.NewCond(&q.lock)
	return q
}

// Put the item in the queue
func (q *ImageQueue) Enqueue(item ImageInfo) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, item)
	q.totalSize += item.Size
	// Cond signals other go routines to execute
	q.cond.Signal()
}

// Gets the item from queue
func (q *ImageQueue) Dequeue() ImageInfo {
	q.lock.Lock()
	defer q.lock.Unlock()
	// if Get is called before Put, then cond waits until the Put signals.
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.totalSize -= item.Size
	return item
}

func (q *ImageQueue) InsertAtFront(items ...ImageInfo) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(items, q.items...)
	for _, item := range items {
		q.totalSize += item.Size
	}
	// Cond signals other go routines to execute
	q.cond.Signal()
}

func (q *ImageQueue) UpdateLastUsed(imageID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	for idx, imgInfo := range q.items {
		if imgInfo.ID == imageID {
			imgInfo.LastUsed = time.Now()
			q.items = append(q.items[:idx], q.items[idx+1:]...)
			q.items = append(q.items, imgInfo)
			break
		}
	}
}

func (q *ImageQueue) Clear() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = []ImageInfo{}
	q.totalSize = 0
}

func (q *ImageQueue) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *ImageQueue) Len() int {
	return len(q.items)
}

func (q *ImageQueue) TotalSize() int64 {
	return q.totalSize
}
