package guac

import (
	"sync"
	"sync/atomic"
)

type CountedLock struct {
	core     sync.Mutex
	numLocks int32
}

func (r *CountedLock) Lock() {
	atomic.AddInt32(&r.numLocks, 1)
	r.core.Lock()
}

func (r *CountedLock) Unlock() {
	atomic.AddInt32(&r.numLocks, -1)
	r.core.Unlock()
}

func (r *CountedLock) HasQueued() bool {
	return atomic.LoadInt32(&r.numLocks) > 1
}
