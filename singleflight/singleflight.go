package singleflight

import (
	"sync"
)

type Call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*Call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*Call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.value, c.err
	}
	c := new(Call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.value, c.err = fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.value, c.err
}
