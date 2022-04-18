package lruk

import (
	"container/list"
	"time"
)

const defaultK = 2

type Value interface {
	Len() int
}

//--------------------------------
// history item
//--------------------------------

type historyItem struct {
	key      string
	views    int
	lastTime time.Time
}

func NewHistoryItem(key string) *historyItem {
	return &historyItem{
		key:      key,
		views:    1,
		lastTime: time.Now(),
	}
}

func (item *historyItem) UpdateTime(d time.Duration) {
	if time.Since(item.lastTime) > d && d != 0 {
		item.views = 1
		item.lastTime = time.Now()
	}
	item.views++
	item.lastTime = time.Now()
}

//--------------------------------
// history view
//--------------------------------

type history struct {
	ll        *list.List
	cache     map[string]*list.Element
	expire    time.Duration
	k         int
	maxBytes  int64
	nBytes    int64
	OnEvicted func(key string)
}

func NewHistory(maxBytes int64, k int, expire time.Duration, onEvicted1 func(key string)) *history {
	return &history{
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		expire:    expire,
		k:         k,
		maxBytes:  maxBytes,
		OnEvicted: onEvicted1,
	}
}

func (h *history) Add(key string, value Value, m *memory) {
	if ele, ok := h.cache[key]; ok {
		h.UpdateHistory(ele, m, value)
	} else {
		ele := h.ll.PushBack(NewHistoryItem(key))
		h.cache[key] = ele
		h.nBytes += int64(len(key))
	}
	for h.maxBytes != 0 && h.maxBytes < h.nBytes {
		h.RemoveOldest()
	}
}

func (h *history) UpdateHistory(ele *list.Element, m *memory, value Value) {
	item := ele.Value.(*historyItem)
	item.UpdateTime(h.expire)
	if item.views >= h.k {
		h.ll.Remove(ele)
		delete(h.cache, item.key)
		h.nBytes -= int64(len(item.key))
		m.Add(item.key, value)
	} else {
		h.ll.MoveToBack(ele)
	}
}

func (h *history) RemoveOldest() {
	ele := h.ll.Front()
	if ele != nil {
		h.ll.Remove(ele)
		item := ele.Value.(*historyItem)
		delete(h.cache, item.key)
		h.nBytes -= int64(len(item.key))
		if h.OnEvicted != nil {
			h.OnEvicted(item.key)
		}
	}
}

//--------------------------------
// memory item
//--------------------------------

type memoryItem struct {
	key      string
	value    Value
	lastTime time.Time
}

func NewMemoryItem(key string, value Value) *memoryItem {
	return &memoryItem{
		key:      key,
		value:    value,
		lastTime: time.Now(),
	}
}

//------------------------------
// LRU-K memory cache
//------------------------------

type memory struct {
	ll        *list.List
	cache     map[string]*list.Element
	maxBytes  int64
	nBytes    int64
	OnEvicted func(key string, value Value)
}

func Newmemory(maxBytes int64, onEvicted func(key string, value Value)) *memory {
	return &memory{
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		maxBytes:  maxBytes,
		OnEvicted: onEvicted,
	}
}

func (m *memory) Get(key string) (Value, bool) {
	if ele, ok := m.cache[key]; ok {
		item := ele.Value.(*memoryItem)
		item.lastTime = time.Now()
		ele.Value = item
		m.ll.MoveToBack(ele)
		return item.value, true
	}
	return nil, false
}

func (m *memory) Add(key string, value Value) {
	if ele, ok := m.cache[key]; ok {
		item := ele.Value.(*memoryItem)
		m.nBytes += int64(value.Len()) - int64(item.value.Len())
		item.lastTime = time.Now()
		item.value = value
		ele.Value = item
		m.ll.MoveToBack(ele)
	} else {
		ele := m.ll.PushBack(NewMemoryItem(key, value))
		m.cache[key] = ele
		m.nBytes += int64(len(key)) + int64(value.Len())
	}
	for m.maxBytes != 0 && m.maxBytes < m.nBytes {
		m.RemoveOldest()
	}
}

func (m *memory) isExit(key string) bool {
	_, ok := m.cache[key]
	return ok
}

func (m *memory) RemoveOldest() {
	ele := m.ll.Front()
	if ele != nil {
		m.ll.Remove(ele)
		item := ele.Value.(*memoryItem)
		delete(m.cache, item.key)
		m.nBytes -= int64(len(item.key)) + int64(item.value.Len())
		if m.OnEvicted != nil {
			m.OnEvicted(item.key, item.value)
		}
	}
}

//------------------------------
// LRU-K memory cache manager
//------------------------------

type Cache struct {
	historyCache *history
	memoryCache  *memory
}

func New(maxBytes int64, k int, expire time.Duration, onEvicted1 func(key string), onEvicted2 func(key string, value Value)) *Cache {
	if k == 0 {
		k = defaultK
	}
	return &Cache{
		historyCache: NewHistory(maxBytes, k, expire, onEvicted1),
		memoryCache:  Newmemory(maxBytes, onEvicted2),
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	m := c.memoryCache
	if value, ok := m.Get(key); ok {
		return value, true
	}
	return nil, false
}

func (c *Cache) Add(key string, value Value) {
	h, m := c.historyCache, c.memoryCache
	if m.isExit(key) {
		m.Add(key, value)
	} else {
		h.Add(key, value, m)
	}
}

func (c *Cache) HLen() int {
	return c.historyCache.ll.Len()
}

func (c *Cache) CLen() int {
	return c.memoryCache.ll.Len()
}
