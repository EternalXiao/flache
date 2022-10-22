package flache

import "sync"

type shard struct {
	lock    sync.Mutex
	indices map[uint64]uint32
	ringBuf *ringBuffer
}

func newShard(size int) *shard {
	s := &shard{
		indices: make(map[uint64]uint32),
		ringBuf: newRingBuffer((size+defaultBlockSize-1)/defaultBlockSize*defaultBlockSize, defaultBlockSize),
	}
	return s
}
