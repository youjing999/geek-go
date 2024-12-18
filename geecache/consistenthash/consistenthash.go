package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	// 采取依赖注入的方式，允许用于替换成自定义的 Hash 函数
	hash Hash
	// 虚拟节点倍数
	replicas int
	// 哈希环
	keys []int // Sorted
	// 虚拟节点与真实节点的映射表, 键是虚拟节点的哈希值, 值是真实节点的名称
	hashMap map[int]string
}

// New creates a map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对每一个真实节点key 创建m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// m.hash() 计算虚拟节点的哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 添加到环上
			m.keys = append(m.keys, hash)
			// 增加虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	// 环上的哈希值排序
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	// 计算 key 的哈希值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	// 如果需要找到的元素确实存在，那么 idx 将是该元素在切片中的索引
	// 如果需要找到的元素不存在，那么 idx 将是该元素应该插入的位置。
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 取模运算 idx % len(m.keys) 来确保索引在有效范围内
	// m.keys[idx % len(m.keys)] 取 m.keys 切片中对应位置的值
	// 将 m.keys 切片中的值作为键，从 m.hashMap 中获取相应的值
	return m.hashMap[m.keys[idx%len(m.keys)]] // 映射得到真实的节点
}
