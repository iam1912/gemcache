package singleflightchan

type request struct {
	key    string
	fn     func() (interface{}, error)
	result chan result
}

type result struct {
	val interface{}
	err error
}

type entry struct {
	res   result
	views int
	ready chan struct{}
}

type Group struct {
	chRequest chan request
	chNum     chan string
	m         map[string]*entry
}

func New() *Group {
	g := &Group{
		chRequest: make(chan request),
		chNum:     make(chan string),
		m:         make(map[string]*entry),
	}
	go g.serve()
	go g.remove()
	return g
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	req := request{
		key:    key,
		fn:     fn,
		result: make(chan result),
	}
	g.chRequest <- req
	result := <-req.result
	return result.val, result.err
}

func (g *Group) remove() {
	for {
		key := <-g.chNum
		e, ok := g.m[key]
		if ok {
			e.views--
			if e.views == 0 {
				delete(g.m, key)
			}
		}
	}
}

func (g *Group) serve() {
	for req := range g.chRequest {
		if e, ok := g.m[req.key]; !ok {
			e := &entry{
				ready: make(chan struct{}),
			}
			g.m[req.key] = e
			go e.call(req, g)
		} else {
			go e.deliver(req, g)
		}
	}
}

func (e *entry) call(req request, g *Group) {
	e.views++
	e.res.val, e.res.err = req.fn()
	req.result <- e.res
	g.chNum <- req.key
	close(e.ready)
}

func (e *entry) deliver(req request, g *Group) {
	e.views++
	<-e.ready
	g.chNum <- req.key
	req.result <- e.res
}
