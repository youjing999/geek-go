package geecache

import (
	"geecache/geecache/lru"
	"sync"
)

type cache struct {
	// 之所以不用读写锁 cache 的 get 和 add 都涉及到写操作(LRU 将最近访问元素移动到链表头)，所以不能直接改为读写锁
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// maxBytes callBack
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
