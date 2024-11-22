package lru

import (
	"container/list"
	"fmt"
)

// entry 是双向链表节点的数据类型，在链表中仍保存每个值对应的key的好处在于，
// 淘汰队首节点时，要用key从字典中删除对应的映射，
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64                         //允许使用的最大内存
	useBytes  int64                         // 当前已使用的内存
	ll        *list.List                    // 标准库实现的双向链表
	cache     map[interface{}]*list.Element //键是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

// New is the constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[interface{}]*list.Element),
		OnEvicted: onEvicted,
	}

}

// Get research
func (c *Cache) Get(key string) (value Value, ok bool) {
	/*
		1.从字典中找到对应的双向链表的节点
		2.将该节点移动到队尾
	*/
	element, ok := c.cache[key]
	// 键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
	if ok {
		// 将链表中的节点 ele 移动到队尾
		// c.ll.MoveToFront(ele)，即将链表中的节点 ele 移动到队尾
		//（双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾）
		c.ll.MoveToFront(element)
		// 类型断言，将entry指针存入，并返回一个*entry类型的值
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 缓存淘汰。即移除最近最少访问的节点
func (c *Cache) RemoveOldest() {
	fmt.Println("call RemoveOldest method")
	// 取到队首节点，从链表中删除
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		kv := element.Value.(*entry)
		// 从字典中 c.cache 删除该节点的映射关系。
		delete(c.cache, kv.key)
		// 更新当前所用的内存
		c.useBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果回调函数 OnEvicted 不为 nil，则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// 如果键存在，则更新对应节点的值，并将该节点移到队尾。
	element, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		c.useBytes += int64(value.Len()) - int64(kv.value.Len())
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.useBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.useBytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
