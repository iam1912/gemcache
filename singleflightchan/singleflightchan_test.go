package singleflightchan

import (
	"sync"
	"testing"
)

func TestSingleflightchan(t *testing.T) {
	g := New()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := g.Do("key1", func() (interface{}, error) {
				return "value1", nil
			})
			if err != nil {
				t.Fatalf("%s", err.Error())
			} else if data.(string) != "value1" {
				t.Fatalf("key1 is error")
			}
		}()
	}
	wg.Wait()
}
