package main

import (
	"flag"
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

func createGroup() *gocache.Group {
	return gocache.NewGroup("id", 2<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[MockDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *gocache.Group) {
	// pool的self = addr
	peers := gocache.NewHttpPool(addr)
	// pool 的 peers = addrs，即一致性hash实例中的 real node
	peers.Set(addrs...)
	// 挂载到group实例的peer字段上
	group.RegisterPeer(peers)

	log.Println("gocache is running at", addr)
	// peers是一个HttpPool，实现了ServeHTTP方法，可以作为handler使用
	// 通过http.get 查询remote节点的缓存值
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			key := req.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool

	// main 函数需要命令行传入 port 和 api 2 个参数，用来在指定端口启动 HTTP 服务。
	flag.IntVar(&port, "port", 8001, "Gocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		// 用来启动一个 API 服务（端口 9999），与用户进行交互，用户感知
		go startAPIServer(apiAddr, group)
	}

	// 用来启动缓存服务器：创建 HTTPPool，添加节点信息，注册到group中
	// 启动 HTTP 服务（共3个端口，8001/8002/8003），用户不感知
	startCacheServer(addrMap[port], []string(addrs), group)

}
