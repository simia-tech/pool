package pool

import "sync/atomic"

type page struct {
	containerSize int
	containers    []*container
}

func (p *page) get(size int) []byte {
	for _, container := range p.containers {
		if container.inUse() {
			continue
		}
		buffer := container.bytes[:size]
		newPointer := pointerOf(buffer)
		if atomic.CompareAndSwapUintptr(&container.pointer, 0, newPointer) {
			return buffer
		}
	}
	return nil
}

func (p *page) countUsedContainers() int {
	count := 0
	for _, container := range p.containers {
		if container.inUse() {
			count++
		}
	}
	return count
}

func (p *page) freeContainer(bufferPtr uintptr) bool {
	for _, container := range p.containers {
		if atomic.CompareAndSwapUintptr(&container.pointer, bufferPtr, 0) {
			return true
		}
	}
	return false
}
