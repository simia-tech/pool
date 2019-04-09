package pool

import (
	"sort"
	"sync"
	"unsafe"
)

// Pool deinfes a byte buffer pool.
type Pool struct {
	Histogram  *Histogram
	pages      []*page
	pagesMutex sync.RWMutex
}

// New returns a new bytes buffer pool with the provided configuration.
func New(configuration Configuration) *Pool {
	p := &Pool{}
	p.Reconfigure(configuration)
	return p
}

// Reconfigure re-initializes the internal containers according
// to the provided configuration.
func (p *Pool) Reconfigure(configuration Configuration) {
	sizes := make([]int, len(configuration))
	index := 0
	for size := range configuration {
		sizes[index] = size
		index++
	}
	sort.Ints(sizes)

	p.pagesMutex.Lock()
	p.pages = make([]*page, len(sizes))
	for i, size := range sizes {
		containerCount := configuration[size]
		containers := make([]*container, containerCount)
		for j := range containers {
			containers[j] = &container{bytes: make([]byte, size)}
		}
		p.pages[i] = &page{
			containerSize: size,
			containers:    containers,
		}
	}
	p.pagesMutex.Unlock()
}

// Get returns a buffer of the provided size.
func (p *Pool) Get(size int) []byte {
	if p.Histogram != nil {
		p.Histogram.Hit(size)
	}

	if size <= 0 {
		return nil
	}

	p.pagesMutex.RLock()
	for index := 0; index < len(p.pages); index++ {
		page := p.pages[index]
		if page.containerSize >= size {
			buffer := page.get(size)
			if buffer != nil {
				p.pagesMutex.RUnlock()
				return buffer
			}
		}
	}
	p.pagesMutex.RUnlock()

	return nil
}

// Put places the provided buffer back in the pool.
func (p *Pool) Put(buffer []byte) error {
	if buffer == nil {
		return ErrNilBuffer
	}
	p.pagesMutex.RLock()
	freed := p.freeContainer(pointerOf(buffer))
	p.pagesMutex.RUnlock()
	if !freed {
		return ErrNoPoolBuffer
	}
	return nil
}

// Usage returns a map with counts of all sizes.
func (p *Pool) Usage() Usage {
	usage := Usage{}
	p.pagesMutex.RLock()
	for _, page := range p.pages {
		usage[page.containerSize] = page.countUsedContainers()
	}
	p.pagesMutex.RUnlock()
	return usage
}

func (p *Pool) freeContainer(bufferPtr uintptr) bool {
	for _, page := range p.pages {
		if page.freeContainer(bufferPtr) {
			return true
		}
	}
	return false
}

func pointerOf(buffer []byte) uintptr {
	return uintptr(unsafe.Pointer(&buffer[0]))
}
