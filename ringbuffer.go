package flache

import "time"

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

func (r *ringBuffer) readVal(index uint32) []byte {

}

func (r *ringBuffer) readKey(index uint32) string {

}

func (r *ringBuffer) readExpireAt(index uint32) time.Time {

}

func (r *ringBuffer) write(e entry) uint32 {

}

func (r *ringBuffer) remove(index uint32) {

}

func (r *ringBuffer) moveToHead(index uint32) {

}
