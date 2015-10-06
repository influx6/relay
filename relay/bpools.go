package relay

import "bytes"

// bufPool represents a reusable buffer pool for executing templates into.
var bufPool = NewBufferPool(1000, 1024)

// BufferPool implements a pool of bytes.Buffers in the form of a bounded channel.
// Pulled from the github.com/oxtoacart/bpool package (Apache licensed).
type BufferPool struct {
	c     chan *bytes.Buffer
	alloc int
}

// NewBufferPool creates a new BufferPool bounded to the given size.
func NewBufferPool(size int, mem int) (bp *BufferPool) {
	return &BufferPool{
		c:     make(chan *bytes.Buffer, size),
		alloc: mem,
	}
}

// Get gets a Buffer from the BufferPool, or creates a new one if none are
// available in the pool.
func (bp *BufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c:
	// reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer(make([]byte, 0, bp.alloc))
	}
	return
}

// Put returns the given Buffer to the BufferPool.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()

	if cap(b.Bytes()) > bp.alloc {
		b = bytes.NewBuffer(make([]byte, 0, bp.alloc))
	}

	select {
	case bp.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}
