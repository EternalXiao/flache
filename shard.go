package flache

import (
	"sync"
	"time"
)

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

func (s *shard) get(key string, hashedKey uint64) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	index, ok := s.indices[hashedKey]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if s.isExpire(index) {
		s.ringBuf.remove(index)
		delete(s.indices, hashedKey)
		return nil, ErrKeyNotFound
	}

	cachedKey := s.ringBuf.readKey(index)
	if key != cachedKey {
		return nil, ErrKeyNotFound
	}

	val := s.ringBuf.readVal(index)
	s.ringBuf.moveToHead(index)
	return val, nil
}

func (s *shard) isExpire(index uint32) bool {
	expireAt := s.ringBuf.readExpireAt(index)
	if expireAt.UnixNano() == 0 {
		return false
	}
	return expireAt.Before(time.Now())
}

func (s *shard) set(key string, hashedKey uint64, value []byte, expiration time.Duration) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	index, ok := s.indices[hashedKey]
	if ok {
		s.ringBuf.remove(index)
	}

	entry := newEntry(key, hashedKey, value, expiration)
	if !s.ringBuf.hasEnoughSpace(entry) {
		return ErrNotEnoughSpace
	}

	for !s.ringBuf.hasEnoughBlocks(entry) {
		tail := s.ringBuf.getTail()
		s.remove(tail)
	}
	index = s.ringBuf.write(entry)
	s.ringBuf.insertHead(index)
	s.indices[hashedKey] = index
	return nil
}

func (s *shard) remove(index uint32) {
	hashedKey := s.ringBuf.readHashedKey(index)
	s.ringBuf.remove(index)
	delete(s.indices, hashedKey)
}

func (s *shard) del(key string, hashedKey uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	index, ok := s.indices[hashedKey]
	if !ok {
		return nil
	}

	cachedKey := s.ringBuf.readKey(index)
	if key != cachedKey {
		return nil
	}

	s.remove(index)
	return nil
}
