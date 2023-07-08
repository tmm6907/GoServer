package caches

import "container/list"

type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
}

type CacheItem struct {
	key   string
	value interface{}
}

var CACHE = NewLRUCache(2048)

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		return elem.Value.(*CacheItem).value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, value interface{}) {
	if elem, ok := c.cache[key]; ok {
		elem.Value.(*CacheItem).value = value
		c.lruList.MoveToFront(elem)
	} else {
		if len(c.cache) >= c.capacity {
			// Evict the least recently used item
			last := c.lruList.Back()
			delete(c.cache, last.Value.(*CacheItem).key)
			c.lruList.Remove(last)
		}

		newElem := c.lruList.PushFront(&CacheItem{key, value})
		c.cache[key] = newElem
	}
}
