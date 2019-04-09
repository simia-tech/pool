package pool

import (
	"sort"
	"sync"
)

// Pool deinfes the byte buffer pool.
type Pool struct {
	sizeLimit  uint64
	size       uint64
	pages      []*page
	pagesMutex sync.Mutex
}

// New returns a new bytes buffer pool with the provided size limit.
func New(l uint64) *Pool {
	return &Pool{sizeLimit: l, size: 0}
}

// Get returns a buffer of the provided size.
func (p *Pool) Get(size uint64) ([]byte, error) {
	p.pagesMutex.Lock()

	// check if a container with at least size bytes is free.
	exactPage := (*page)(nil)
	for index := 0; index < len(p.pages); index++ {
		page := p.pages[index]

		if page.containerSize >= size {
			buffer := page.get(size)
			if buffer != nil {
				p.pagesMutex.Unlock()
				return buffer, nil
			}
		}

		if page.containerSize == size {
			exactPage = page
		}
	}

	// at this point a new container will be created. before that happens, the size limit is checked.
	if p.size+size > p.sizeLimit {
		p.pagesMutex.Unlock()
		return nil, ErrPoolLimitReached
	}
	p.size += size

	// if no fitting free container exists and no page with the exact size exists, create one.
	// containing a single container.
	if exactPage == nil {
		page := newPage(size)
		p.pages = append(p.pages, page)
		p.sortPages()
		buffer := page.addContainer()
		p.pagesMutex.Unlock()
		return buffer, nil
	}

	// if a page with the exact size exists, add a new container.
	buffer := exactPage.addContainer()

	p.pagesMutex.Unlock()
	return buffer, nil
}

// Put places the provided buffer back in the pool.
func (p *Pool) Put(buffer []byte) error {
	if buffer == nil {
		return ErrNilBuffer
	}

	p.pagesMutex.Lock()
	for _, page := range p.pages {
		if page.freeContainer(buffer) {
			p.pagesMutex.Unlock()
			return nil
		}
	}
	p.pagesMutex.Unlock()

	return ErrNoPoolBuffer
}

// Size returns the size of the byte buffer pool.
func (p *Pool) Size() uint64 {
	return p.size
}

// Usage returns a map with counts of all sizes.
func (p *Pool) Usage() Usage {
	usage := Usage{}
	p.pagesMutex.Lock()
	for _, page := range p.pages {
		usage[page.containerSize] = page.usageEntry()
	}
	p.pagesMutex.Unlock()
	return usage
}

func (p *Pool) sortPages() {
	sort.Slice(p.pages, func(i int, j int) bool {
		return p.pages[i].containerSize < p.pages[j].containerSize
	})
}
