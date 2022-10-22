package flache

type Cache struct {
	shards    []*shard
	shardMask int
}

func NewCache() *Cache {
	// TODO: use conf
	c := &Cache{
		shards:    make([]*shard, defaultShardSize),
		shardMask: defaultShardSize - 1,
	}
	for i := range c.shards {
		c.shards[i] = newShard(defaultCacheSize / defaultShardSize)
	}
	return c
}
