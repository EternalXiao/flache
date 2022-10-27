package flache

import (
	"encoding/binary"
	"time"
)

type entry []byte

func newEntry(key string, hashedKey uint64, value []byte, expiration time.Duration) entry {
	buf := make([]byte, entryHeaderSize+len(key)+len(value))
	binary.LittleEndian.PutUint16(buf[keyLengthOffset:], uint16(len(key)))
	binary.LittleEndian.PutUint32(buf[valueLengthOffset:], uint32(len(value)))
	if expiration != 0 {
		binary.LittleEndian.PutUint64(buf[expireAtOffset:], uint64(time.Now().Add(expiration).UnixNano()))
	}
	binary.LittleEndian.PutUint64(buf[hashedKeyOffset:], hashedKey)
	return buf
}
