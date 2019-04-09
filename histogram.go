package pool

import (
	"sync"
	"sync/atomic"
)

// Histogram defines a size histogram.
type Histogram struct {
	data sync.Map
}

// NewHistogram returns a new histogram.
func NewHistogram() *Histogram {
	return &Histogram{}
}

// Hit records a hit for the provided size.
func (h *Histogram) Hit(size int) {
	count := uint64(0)
	countPtr, _ := h.data.LoadOrStore(size, &count)
	atomic.AddUint64(countPtr.(*uint64), 1)
}

// Map returns a map containing the histogram data.
func (h *Histogram) Map() map[int]uint64 {
	result := map[int]uint64{}
	h.data.Range(func(key interface{}, value interface{}) bool {
		result[key.(int)] = atomic.LoadUint64(value.(*uint64))
		return true
	})
	return result
}
