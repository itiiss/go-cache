package go_cache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 实现了Getter接口
// 函数类型实现某一个接口，称之为接口型函数，
// 使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。
type GetterFunc func(key string) ([]byte, error)

// Get 为GetterFunc关联Get方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 name
type Group struct {
	name      string // namespace
	getter    Getter // if cache miss, execute getter callback
	mainCache cache  // LRU cache instance
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func getGroup(name string) *Group {
	//使用了只读锁 RLock()，因为不涉及任何冲突变量的写操作。
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}
	v, ok := g.mainCache.Get(key)
	if ok {
		log.Println("[go-cache] hit")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

// 通过Getter方法从本地节点获取数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	// 从 Getter中取到数据后，更新到mainCache中
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)

	return value, nil
}

// 添加 entry 到 mainCache实例
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
