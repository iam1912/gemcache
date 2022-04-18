package gemcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPPool(t *testing.T) {
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPool(addr)

	testCases := map[string]string{
		"630":               "http://localhost:9999/_gemcache/scores/Tom",
		"unknown not exist": "http://localhost:9999/_gemcache/scores/unknown",
	}
	for expected, addr := range testCases {
		data := performRequest(peers, "GET", addr)
		if strings.TrimSpace(data) != strings.TrimSpace(expected) {
			t.Fatalf("%s expected result is %s but got %s", addr, expected, data)
		}
	}
}

func performRequest(r http.Handler, method string, url string) string {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, nil)
	r.ServeHTTP(w, req)
	return w.Body.String()
}
