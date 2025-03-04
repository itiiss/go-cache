package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (e *entry) sizeInt64() int64 {
	return int64(len(e.key)) + int64(e.value.Len())
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

func (c *Cache) Get(key string) (Value, bool) {
	ele, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	// 取出最旧元素
	ele := c.ll.Back()
	if ele != nil {
		// 从 list中删除元素
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		//取出元素的key，在map中也删除该记录
		delete(c.cache, kv.key)
		// 更新当前已用的cache size
		c.nBytes -= kv.sizeInt64()
		// 执行回掉函数
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	ele, ok := c.cache[key]
	if ok {
		// map中存在key，将其移动位置和更新value，最后更新size
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.value = value
		c.nBytes += kv.sizeInt64()
	} else {
		// 尚未存在该entry，新建entry到队首，并在map中建立映射
		entry := &entry{key, value}
		ele := c.ll.PushFront(entry)
		c.cache[key] = ele
		c.nBytes += entry.sizeInt64()
	}
	// 超出cache大小时，持续移除最旧的
	for c.maxBytes > 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}
