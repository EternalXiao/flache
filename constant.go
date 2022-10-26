package flache

import "errors"

const (
	defaultShardSize = 16
	defaultBlockSize = 1024
	defaultCacheSize = defaultShardSize * defaultBlockSize * 1024 * 64
)

var ErrKeyNotFount = errors.New("key not found")
