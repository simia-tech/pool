package pool_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simia-tech/pool"
)

func TestPoolGet(t *testing.T) {
	tcs := []struct {
		name             string
		configuration    pool.Configuration
		size             int
		expectBuffer     bool
		expectBufferSize int
		expectUsage      pool.Usage
	}{
		{"Simple", pool.Configuration{10: 1}, 10, true, 10, pool.Usage{10: 1}},
		{"Smaller", pool.Configuration{10: 1}, 8, true, 8, pool.Usage{10: 1}},
		{"TooLarge", pool.Configuration{10: 1}, 12, false, 0, pool.Usage{10: 0}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			p := pool.New(tc.configuration)

			buffer := p.Get(tc.size)
			if tc.expectBuffer {
				require.NotNil(t, buffer)
				assert.Len(t, buffer, tc.expectBufferSize)
			} else {
				assert.Nil(t, buffer)
			}
			assert.Equal(t, tc.expectUsage, p.Usage())
		})
	}

	t.Run("Upgrade", func(t *testing.T) {
		p := pool.New(pool.Configuration{10: 1, 20: 1})

		buffer := p.Get(10)
		require.NotNil(t, buffer)

		buffer = p.Get(12)
		require.NotNil(t, buffer)
		assert.Len(t, buffer, 12)

		assert.Equal(t, pool.Usage{10: 1, 20: 1}, p.Usage())
	})
}

func TestPoolGetAndPut(t *testing.T) {
	p := pool.New(pool.Configuration{10: 1})

	buffer := p.Get(10)
	require.NotNil(t, buffer)
	require.NoError(t, p.Put(buffer))

	assert.Equal(t, pool.Usage{10: 0}, p.Usage())
}

func TestPoolPutError(t *testing.T) {
	p := pool.New(pool.Configuration{10: 1})

	err := p.Put([]byte{1, 2, 3})
	assert.Equal(t, pool.ErrNoPoolBuffer, err)
}

func TestPoolConcurrentReconfiguration(t *testing.T) {
	p := pool.New(pool.Configuration{10: 10})

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	for index := 0; index < 10; index++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				default:
					buffer := p.Get(10)
					require.NoError(t, p.Put(buffer))
				}
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	p.Reconfigure(pool.Configuration{10: 5, 20: 5})
	time.Sleep(100 * time.Millisecond)
	cancel()

	wg.Wait()
}

func BenchmarkPoolGetAndPut(b *testing.B) {
	p := pool.New(pool.Configuration{10: 1})

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		buffer := p.Get(10)
		p.Put(buffer)
	}
	b.StopTimer()
}

func BenchmarkPoolConcurrentGetAndPut(b *testing.B) {
	p := pool.New(pool.Configuration{10: 2})

	wg := sync.WaitGroup{}
	wg.Add(2)

	worker := func() {
		for index := 0; index < b.N; index++ {
			buffer := p.Get(10)
			if buffer == nil {
				log.Printf("buffer = nil")
				return
			}
			require.NotNil(b, buffer)
			p.Put(buffer)
		}
		wg.Done()
	}

	b.ResetTimer()
	go worker()
	go worker()
	wg.Wait()
}
