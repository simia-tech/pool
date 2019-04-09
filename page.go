package pool

import "sync/atomic"

type page struct {
	containerSize uint64
	containers    []*container
}

func newPage(containerSize uint64) *page {
	return &page{containerSize: containerSize}
}

func (p *page) addContainer() []byte {
	container := newContainer(p.containerSize)
	p.containers = append(p.containers, container)
	return container.allocate(p.containerSize)
}

func (p *page) get(size uint64) []byte {
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

func (p *page) freeContainer(buffer []byte) bool {
	for _, container := range p.containers {
		if container.free(buffer) {
			return true
		}
	}
	return false
}

func (p *page) usageEntry() Entry {
	used := 0
	for _, container := range p.containers {
		if container.inUse() {
			used++
		}
	}

	return Entry{
		Used:  used,
		Count: len(p.containers),
	}
}
