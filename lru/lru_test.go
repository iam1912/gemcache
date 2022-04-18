package lru

import (
	"testing"
)

type myString string

func (m myString) Len() int {
	return len(m)
}

func TestLruGet(t *testing.T) {
	c := New(int64(0), nil)
	c.Add("test1", myString("test1"))
	if v, ok := c.Get("test1"); !ok || string(v.(myString)) != "test1" {
		t.Fatalf("cache hit test1=test1 failed")
	}
	if _, ok := c.Get("test2"); ok {
		t.Fatalf("cache miss test2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	keys := []string{"k1", "k2", "k3"}
	values := []string{"v1", "v2", "v3"}
	cap := len(keys[0] + keys[1] + values[0] + values[1])
	c := New(int64(cap), nil)
	for i := 0; i < 3; i++ {
		c.Add(keys[i], myString(values[i]))
	}
	if _, ok := c.Get("k1"); ok || c.Len() != 2 {
		t.Fatalf("removeoldest key1 failed")
	}
}
