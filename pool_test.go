package pool_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simia-tech/pool"
)

func TestPoolGet(t *testing.T) {
	tcs := []struct {
		name        string
		limit       uint64
		size        uint64
		expectErr   error
		expectUsage pool.Usage
	}{
		{"Simple", 100, 10, nil, pool.Usage{10: {Used: 1, Count: 1}}},
		{"ReachLimit", 100, 110, pool.ErrPoolLimitReached, pool.Usage{}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			p := pool.New(tc.limit)

			buffer, err := p.Get(tc.size)
			if tc.expectErr == nil {
				require.NoError(t, err)
				assert.Len(t, buffer, int(tc.size))
			} else {
				assert.Equal(t, tc.expectErr, err)
			}
			assert.Equal(t, tc.expectUsage, p.Usage())
		})
	}
}

func TestPoolPutError(t *testing.T) {
	p := pool.New(100)

	err := p.Put([]byte{1, 2, 3})
	assert.Equal(t, pool.ErrNoPoolBuffer, err)
}

func TestPoolGetAndPut(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		p := pool.New(100)

		buffer, err := p.Get(10)
		require.NoError(t, err)
		require.NotNil(t, buffer)
		require.NoError(t, p.Put(buffer))

		assert.Equal(t, pool.Usage{10: {Used: 0, Count: 1}}, p.Usage())
	})
	t.Run("UseOfBiggerContainer", func(t *testing.T) {
		p := pool.New(100)

		buffer, err := p.Get(10)
		require.NoError(t, err)
		require.NoError(t, p.Put(buffer))

		_, err = p.Get(8)
		require.NoError(t, err)

		assert.Equal(t, pool.Usage{10: {Used: 1, Count: 1}}, p.Usage())
	})
	t.Run("ContainerCreation", func(t *testing.T) {
		p := pool.New(100)

		buffer, err := p.Get(10)
		require.NoError(t, err)
		require.NoError(t, p.Put(buffer))

		_, err = p.Get(12)
		require.NoError(t, err)

		assert.Equal(t, pool.Usage{10: {Used: 0, Count: 1}, 12: {Used: 1, Count: 1}}, p.Usage())
	})
}

func TestPoolSize(t *testing.T) {
	p := pool.New(100)

	p.Get(10)
	p.Get(32)

	assert.Equal(t, uint64(42), p.Size())
}

func BenchmarkPoolGetAndPut(b *testing.B) {
	p := pool.New(100)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		buffer, _ := p.Get(10)
		p.Put(buffer)
	}
	b.StopTimer()
}

func BenchmarkPoolConcurrentGetAndPut(b *testing.B) {
	p := pool.New(100)

	wg := sync.WaitGroup{}
	wg.Add(2)

	worker := func() {
		for index := 0; index < b.N; index++ {
			buffer, err := p.Get(10)
			if err != nil {
				b.Errorf("unexpected error: %v", err)
				return
			}
			p.Put(buffer)
		}
		wg.Done()
	}

	b.ResetTimer()
	go worker()
	go worker()
	wg.Wait()
}
