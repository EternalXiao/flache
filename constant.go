package flache

import "errors"

const (
	defaultShardSize = 16
	defaultBlockSize = 1024
	defaultCacheSize = defaultShardSize * defaultBlockSize * 1024 * 64

	blockHeaderSize   = 4
	entryHeaderSize   = 30
	nextEntryOffset   = 0
	prevEntryOffset   = 4
	keyLengthOffset   = 8
	valueLengthOffset = 10
	expireAtOffset    = 14
	hashedKeyOffset   = 22
)

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrNotEnoughSpace = errors.New("not enough space")
)
