package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 负责与外部交互，控制缓存存储的主流程

// Getter 根据某个键加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 避免了为每个实现创建单独的结构体
type GetterFunc func(key string) ([]byte, error)

// Get 实现get方法
// 函数类型实现某一个接口，称之为接口型函数
//方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//	Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 name。
//	比如可以创建三个 Group，缓存学生的成绩命名为 scores，缓存学生信息的命名为 info，缓存学生课程的命名为 courses。
type Group struct {
	name string
	// getter Getter，即缓存未命中时获取源数据的回调(callback)。
	getter Getter
	// mainCache cache，即一开始实现的并发缓存。
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()

	defer mu.Unlock()

	g := &Group{
		// 缓存的命名空间, 每个 Group 拥有一个唯一的名称 name
		name: name,
		// 缓存未命中时获取源数据的回调
		getter: getter,
		// 并发缓存
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 返回先前使用 NewGroup 创建的命名组，如果不存在这样的组，则返回 nil
func GetGroup(name string) *Group {
	// 只读锁，读的本质也是并发的操作map，如果不加锁，并发量大可能会panic
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	/*
		1. 从 mainCache 中查找缓存，如果存在则返回缓存值
		2. 缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取）
	*/
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	value, ok := g.mainCache.get(key)
	if ok {
		log.Println("[GeeCache hit]")
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{
		b: cloneBytes(bytes),
	}
	g.populateCache(key, value)
	return value, nil

}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
