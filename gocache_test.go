package go_cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {

	var toByteFunc = func(key string) ([]byte, error) {
		return []byte(key), nil
	}

	// f 是个高阶函数，f(x) = toByteFunc(x)
	var f = GetterFunc(toByteFunc)
	// g 和 f是等效的，回调函数转换成了接口 g Getter
	var g Getter = GetterFunc(toByteFunc)

	expect := []byte("key")
	// f.Get("key") 等效于 toByteFunc("key")
	v, _ := f.Get("key")
	if !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed expect %v, got %v", expect, v)
	}

	v, _ = g.Get("key")
	if !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed expect %v, got %v", expect, v)
	}
}

var db = map[string]string{
	"Alice": "111",
	"Bob":   "222",
	"Carl":  "333",
}

func TestGet(t *testing.T) {
	loadCounter := make(map[string]int, len(db))
	g := NewGroup("id", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[MockDB] search key", key)
			// mockDB 中存在 key
			if v, ok := db[key]; ok {
				// loadCounter中不存在该key,初始化该entry，k = key，v = 1
				if _, ok := loadCounter[key]; !ok {
					loadCounter[key] = 1
				} else {
					// loadCounter中存在该key，v +=1
					loadCounter[key] += 1
				}
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))

	for k, v := range db {
		// 第一次Get(k)，此时LRU cache为空
		// 从 GetterFunc取到期望的v，loadCount被设置为1
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)

		}

		// 第二次 Get(k)，期望loadCount仍然为1
		// 这次没有调用GetterFunc，而是直接从LRU cache中取得
		if _, err := g.Get(k); err != nil || loadCounter[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
