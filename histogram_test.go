package pool_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/simia-tech/pool"
)

func TestHistogramHit(t *testing.T) {
	h := pool.NewHistogram()

	h.Hit(10)
	h.Hit(10)
	h.Hit(20)

	assert.Equal(t, map[int]uint64{10: 2, 20: 1}, h.Map())
}
