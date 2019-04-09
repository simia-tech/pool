package pool

import (
	"sync/atomic"
	"unsafe"
)

type container struct {
	bytes   []byte
	pointer uintptr
}

func newContainer(size uint64) *container {
	return &container{
		bytes:   make([]byte, size),
		pointer: 0,
	}
}

func (c *container) inUse() bool {
	return atomic.LoadUintptr(&c.pointer) != 0
}

func (c *container) allocate(size uint64) []byte {
	buffer := c.bytes[:size]
	if !atomic.CompareAndSwapUintptr(&c.pointer, 0, pointerOf(buffer)) {
		panic("container already allocated")
	}
	return buffer
}

func (c *container) free(buffer []byte) bool {
	return atomic.CompareAndSwapUintptr(&c.pointer, pointerOf(buffer), 0)
}

func pointerOf(buffer []byte) uintptr {
	return uintptr(unsafe.Pointer(&buffer[0]))
}
