package flache

type ringBuffer struct {
	buf        []byte
	freeBlocks []int
	capacity   int
	blockSize  int
}

func newRingBuffer(size, blockSize int) *ringBuffer {
	ringBuf := &ringBuffer{
		buf:       make([]byte, size),
		capacity:  size,
		blockSize: blockSize,
	}
	ringBuf.freeBlocks = make([]int, size/blockSize-1)
	for i := range ringBuf.freeBlocks {
		ringBuf.freeBlocks[i] = i + 1
	}
	return ringBuf
}
