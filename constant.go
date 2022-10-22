package flache

const (
	defaultShardSize = 16
	defaultBlockSize = 1024
	defaultCacheSize = defaultShardSize * defaultBlockSize * 1024 * 64
)
