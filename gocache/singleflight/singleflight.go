package singleflight

import "sync"

// 表示正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 主数据结构，管理不同 key 的请求(call)
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 对于每个函数，加上一个key标识，通过key来判断是否再次调用
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// goroutine对map的访问需要是互斥的
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 检查到key已存在，说明已经有其他goroutine在执行key对应的fn
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // WaitGroup 阻塞，停止发出请求，直到锁被释放
		return c.val, c.err
	}

	// 某个key的call第一次进入
	c := new(call)
	c.wg.Add(1) // WaitGroup 锁+1，表示有一个新的请求正在处理
	g.m[key] = c
	g.mu.Unlock()

	// 执行 fn 得到结果
	c.val, c.err = fn()
	c.wg.Done() // 对应add，使WaitGroup 锁-1，幻想其他等待该key的goroutine

	// 执行后删除该key
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
