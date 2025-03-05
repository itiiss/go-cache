package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            // virtual nodes count
	keys     []int          // real node hash value ring
	hashMap  map[int]string // key = virtual node, value = real node
}

func New(replicas int, hashFn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hashFn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 允许传入多个节点名
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对于每一个节点，创建 replica个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点名称为 strconv.Itoa(i) + key，得到其hash
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 添加虚拟节点到环上
			m.keys = append(m.keys, hash)
			// 建立虚拟节点到真实节点的映射
			m.hashMap[hash] = key
		}
	}
	// 对环上的hash值排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 计算得到key的hash值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个value > hash值的节点下标
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//得到该下表对应的虚拟节点，再找到对应的真实节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
