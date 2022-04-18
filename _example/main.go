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
	"fmt"
	"sync"

	"github.com/iam1912/gemcache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func handleRpc() {
	gemcache.HandleService("localhost:9999", "")
}

func main() {
	gemcache.NewGroup("scores", 2<<10, gemcache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	go handleRpc()
	wg := &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := gemcache.GemcacheClientReq("localhost:9999", "scores", "Tom")
			fmt.Println(string(data), err)
		}()
	}
	wg.Wait()
}
