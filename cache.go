package flache

import (
	"time"

	"github.com/cespare/xxhash"
)

type Cache struct {
	shards    []*shard
	shardMask uint64
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

func (c *Cache) Get(key string) ([]byte, error) {
	hashedKey := xxhash.Sum64String(key)
	shard := c.getShard(hashedKey)
	return shard.get(key, hashedKey)
}

func (c *Cache) Set(key string, value []byte, expiration time.Duration) error {
	hashedKey := xxhash.Sum64String(key)
	shard := c.getShard(hashedKey)
	return shard.set(key, hashedKey, value, expiration)
}

func (c *Cache) Del(key string) error {
	hashedKey := xxhash.Sum64String(key)
	shard := c.getShard(hashedKey)
	return shard.del(key, hashedKey)
}

func (c *Cache) getShard(hashedKey uint64) *shard {
	return c.shards[hashedKey&c.shardMask]
}
