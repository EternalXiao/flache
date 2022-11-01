package flache

import (
	"encoding/binary"
	"time"
)

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

func (r *ringBuffer) readPrevEntryIndex(index uint32) uint32 {
	return binary.LittleEndian.Uint32(r.buf[int(index)*r.blockSize+prevEntryOffset:])
}

func (r *ringBuffer) readNextEntryIndex(index uint32) uint32 {
	return binary.LittleEndian.Uint32(r.buf[int(index)*r.blockSize+nextEntryOffset:])
}

func (r *ringBuffer) readVal(index uint32) []byte {
	keyLength := int(r.readKeyLength(index))
	valLength := int(r.readValLength(index))
	val := make([]byte, valLength)
	valStartBlockOffset := 0
	for i := 0; keyLength >= 0; i++ {
		if i == 0 {
			keyLength -= (r.blockSize - blockHeaderSize - entryHeaderSize)
		} else {
			keyLength -= (r.blockSize - blockHeaderSize)
		}
		if keyLength >= 0 {
			index = r.readNextBlockIndex(index)
		} else {
			valStartBlockOffset = r.blockSize + keyLength
		}
	}

	for i, offset := 0, 0; offset < valLength; i++ {
		if i == 0 {
			copy(val[offset:], r.buf[int(index)*r.blockSize+valStartBlockOffset:int(index+1)*r.blockSize])
			offset += r.blockSize - valStartBlockOffset
		} else {
			copy(val[offset:], r.buf[int(index)*r.blockSize+blockHeaderSize:int(index+1)*r.blockSize])
			offset += r.blockSize - blockHeaderSize
		}
		index = r.readNextBlockIndex(index)
	}
	return val
}

func (r *ringBuffer) readKey(index uint32) string {
	keyLength := int(r.readKeyLength(index))
	keyBuf := make([]byte, keyLength)
	for i, offset := 0, 0; offset < keyLength; i++ {
		if i == 0 {
			copy(keyBuf[offset:], r.buf[int(index)*r.blockSize+blockHeaderSize+entryHeaderSize:int(index+1)*r.blockSize])
			offset += r.blockSize - blockHeaderSize - entryHeaderSize
		} else {
			copy(keyBuf[offset:], r.buf[int(index)*r.blockSize+blockHeaderSize:int(index+1)*r.blockSize])
			offset += r.blockSize - blockHeaderSize
		}
		index = r.readNextBlockIndex(index)
	}
	return bytesToString(keyBuf)
}

func (r *ringBuffer) readKeyLength(index uint32) uint16 {
	return binary.LittleEndian.Uint16(r.buf[int(index)*r.blockSize+blockHeaderSize+keyLengthOffset:])
}

func (r *ringBuffer) readValLength(index uint32) uint32 {
	return binary.LittleEndian.Uint32(r.buf[int(index)*r.blockSize+blockHeaderSize+valueLengthOffset:])
}

func (r *ringBuffer) readHashedKey(index uint32) uint64 {
	return binary.LittleEndian.Uint64(r.buf[int(index)*r.blockSize+blockHeaderSize+hashedKeyOffset:])
}

func (r *ringBuffer) readExpireAt(index uint32) time.Time {
	expireAtTs := binary.LittleEndian.Uint64(r.buf[int(index)*r.blockSize+blockHeaderSize+expireAtOffset:])
	return time.Unix(int64(expireAtTs), 0)
}

func (r *ringBuffer) readNextBlockIndex(index uint32) uint32 {
	return binary.LittleEndian.Uint32(r.buf)
}

func (r *ringBuffer) write(e entry) uint32 {
	requiredBlocks := r.calBlocks(e)
	blocks := r.freeBlocks[len(r.freeBlocks)-requiredBlocks:]
	r.freeBlocks = r.freeBlocks[:len(r.freeBlocks)-requiredBlocks]

	for i := 0; i < len(blocks); i++ {
		if i < len(blocks)-1 {
			r.writeNextBlockIndex(uint32(blocks[i]), uint32(blocks[i+1]))
		}
		copy(r.buf[blocks[i]*r.blockSize+blockHeaderSize:(blocks[i]+1)*r.blockSize], e[i*(r.blockSize-blockHeaderSize):])
	}
	return uint32(blocks[0])
}

func (r *ringBuffer) writeNextBlockIndex(index, next uint32) {
	binary.LittleEndian.PutUint32(r.buf[int(index)*r.blockSize:], next)
}

func (r *ringBuffer) writeNextEntryIndex(prev, next uint32) {
	binary.LittleEndian.PutUint32(r.buf[int(prev)*r.blockSize+nextEntryOffset:], next)
}

func (r *ringBuffer) writePrevEntryIndex(next, prev uint32) {
	binary.LittleEndian.PutUint32(r.buf[int(next)*r.blockSize+prevEntryOffset:], prev)
}

func (r *ringBuffer) remove(index uint32) {
	prev := r.readPrevEntryIndex(index)
	next := r.readPrevEntryIndex(index)
	r.link(prev, next)

	r.freeBlocks = append(r.freeBlocks, int(index))
	nextBlock := r.readNextBlockIndex(index)
	for ; nextBlock != 0; nextBlock = r.readNextBlockIndex(index) {
		index = nextBlock
		r.freeBlocks = append(r.freeBlocks, int(index))
	}
}

func (r *ringBuffer) insertHead(index uint32) {
	head := r.getHead()
	r.link(index, head)
	r.link(0, index)
}

func (r *ringBuffer) moveToHead(index uint32) {
	prev := r.readPrevEntryIndex(index)
	next := r.readPrevEntryIndex(index)
	r.link(prev, next)
	r.insertHead(index)
}

func (r *ringBuffer) link(prev, next uint32) {
	r.writeNextEntryIndex(prev, next)
	r.writePrevEntryIndex(next, prev)
}

func (r *ringBuffer) getHead() uint32 {
	return binary.LittleEndian.Uint32(r.buf[blockHeaderSize+nextEntryOffset:])
}

func (r *ringBuffer) getTail() uint32 {
	return binary.LittleEndian.Uint32(r.buf[blockHeaderSize+prevEntryOffset:])
}

func (r *ringBuffer) calBlocks(e entry) int {
	return (len(e) + r.blockSize - blockHeaderSize - 1) / (r.blockSize - blockHeaderSize)
}

func (r *ringBuffer) hasEnoughSpace(e entry) bool {
	return (r.capacity/r.blockSize - 1) >= r.calBlocks(e)
}

func (r *ringBuffer) hasEnoughBlocks(e entry) bool {
	return r.calBlocks(e) >= len(r.freeBlocks)
}
