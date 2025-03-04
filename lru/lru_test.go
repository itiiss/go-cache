package lru

import (
	"log"
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lruCache := NewCache(int64(0), nil)
	lruCache.Add("key1", String("value1"))
	v, ok := lruCache.Get("key1")
	if !ok || string(v.(String)) != "value1" {
		t.Fatal("Get hit key1=value1 fail")
	}
	_, ok = lruCache.Get("key2")
	if ok {
		t.Fatal("Get miss key2 fail")
	}
}

func TestRemove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "v1", "v2", "v3"

	cap := len(k1 + v1 + k2 + v2)
	lruCache := NewCache(int64(cap), nil)

	lruCache.Add(k1, String(v1))
	lruCache.Add(k2, String(v2))
	lruCache.Add(k3, String(v3))

	_, ok := lruCache.Get("key1")

	log.Println(ok, lruCache.Len())

	if ok || lruCache.Len() != 2 {
		t.Fatal("Remove oldest key1=v1 fail")
	}
}

func TestOnEvicted(t *testing.T) {
	// 维护所有移除的keys
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lruCache := NewCache(int64(10), callback)

	lruCache.Add("key1", String("123456"))
	lruCache.Add("k2", String("k2"))
	lruCache.Add("k3", String("k3"))
	lruCache.Add("k4", String("k4"))

	// key1和k2 需要被移除
	expected := []string{"key1", "k2"}

	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expected)
	}

}
