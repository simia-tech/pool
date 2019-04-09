package pool

// Usage defines a container map where usage information about the buffer usage is stored.
type Usage map[uint64]Entry

// Entry defines a usage entry.
type Entry struct {
	Used  int
	Count int
}
