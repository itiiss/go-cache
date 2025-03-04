package main

import (
	"fmt"
	"gocache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Alice": "111",
	"Bob":   "222",
	"Carl":  "333",
}

func main() {
	gocache.NewGroup("id", 2<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[MockDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key %s not found", key)
		}))

	addr := "localhost:9999"
	peers := gocache.NewHttpPool(addr)
	log.Println("go_cache is running on", addr)
	log.Fatal(http.ListenAndServe(addr, peers))

}
