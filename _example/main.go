// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/iam1912/gemcache"
// )

// var db = map[string]string{
// 	"Tom":  "630",
// 	"Jack": "589",
// 	"Sam":  "567",
// }

// func main() {
// 	gemcache.NewGroup("scores", 2<<10, gemcache.GetterFunc(
// 		func(key string) ([]byte, error) {
// 			if v, ok := db[key]; ok {
// 				return []byte(v), nil
// 			}
// 			return nil, fmt.Errorf("%s not exist", key)
// 		}
// 	))
// 	addr := "localhost:9999"
// 	peers := gemcache.NewHTTPool(addr)
// 	log.Println("gemcache is running at", addr)
// 	log.Fatal(http.ListenAndServe(addr, peers))
// }

//gemRpc
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/iam1912/gemcache"
)

// import (
// 	"fmt"
// 	"sync"

// 	"github.com/iam1912/gemcache"
// )

// var db = map[string]string{
// 	"Tom":  "630",
// 	"Jack": "589",
// 	"Sam":  "567",
// }

// func handleRpc() {
// 	gemcache.HandleService("localhost:9999", "")
// }

// func main() {
// 	gemcache.NewGroup("scores", 2<<10, gemcache.GetterFunc(
// 		func(key string) ([]byte, error) {
// 			if v, ok := db[key]; ok {
// 				return []byte(v), nil
// 			}
// 			return nil, fmt.Errorf("%s not exist", key)
// 		}))
// 	go handleRpc()
// 	wg := &sync.WaitGroup{}
// 	for i := 0; i < 2; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			data, err := gemcache.GemcacheClientReq("localhost:9999", "scores", "Tom")
// 			fmt.Println(string(data), err)
// 		}()
// 	}
// 	wg.Wait()
// }

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gemcache.Group {
	return gemcache.NewGroup("scores", 2<<10, gemcache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gem *gemcache.Group) {
	peers := gemcache.NewHTTPool(addr)
	peers.Set(addrs...)
	gem.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gem *gemcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gem.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
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

	gem := createGroup()
	if api {
		go startAPIServer(apiAddr, gem)
	}
	startCacheServer(addrMap[port], []string(addrs), gem)
}
