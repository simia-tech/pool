package pool

import "sync/atomic"

type container struct {
	bytes   []byte
	pointer uintptr
}

func (c *container) inUse() bool {
	return atomic.LoadUintptr(&c.pointer) != 0
}
