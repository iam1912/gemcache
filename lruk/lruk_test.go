package lruk

import (
	"testing"
	"time"
)

type myString string

func (m myString) Len() int {
	return len(m)
}

func TestLruGet(t *testing.T) {
	c := New(int64(0), 2, time.Second*30, nil, nil)
	c.Add("k1", myString("v1"))
	c.Add("k1", myString("v11"))
	c.Add("k1", myString("v111"))

	c.Add("k2", myString("v2"))
	if v, ok := c.Get("k1"); !ok || string(v.(myString)) != "v111" {
		t.Fatalf("cache hit test1=test1 failed")
	}
	if _, ok := c.Get("k2"); ok {
		t.Fatalf("cache miss test2 failed")
	}
}

func TestHistoryRemoveOldest(t *testing.T) {
	keys := []string{"k1", "k2", "k3"}
	values := []string{"v1", "v2", "v3"}
	cap := len(keys[0] + keys[1])

	c := New(int64(cap), 5, time.Second*30, nil, nil)
	for i := 0; i < 3; i++ {
		c.Add(keys[i], myString(values[i]))
	}
	if _, ok := c.Get("k1"); ok || c.HLen() != 2 {
		t.Fatalf("history removeoldest key1 failed")
	}
}

func TestCacheRemoveOldest(t *testing.T) {
	keys := []string{"k1", "k2", "k3"}
	values := []string{"v1", "v2", "v3"}
	cap := len(keys[0] + keys[1] + values[0] + values[1])
	c := New(int64(cap), 2, time.Second*30, nil, nil)
	for i := 0; i < 3; i++ {
		c.Add(keys[i], myString(values[i]))
		c.Add(keys[i], myString(values[i]))
	}
	if _, ok := c.Get("k1"); ok || c.CLen() != 2 {
		t.Fatalf("removeoldest key1 failed")
	}
}
