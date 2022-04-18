package consistenthashmap

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

type HashRing []uint32

func (r HashRing) Len() int           { return len(r) }
func (r HashRing) Less(i, j int) bool { return r[i] < r[j] }
func (r HashRing) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

type Map struct {
	hash     Hash
	replicas int
	keys     HashRing
	nums     map[string]int
	ring     map[uint32]string
	mu       sync.RWMutex
}

func New(replicas int, fn Hash) *Map {
	h := &Map{
		hash:     fn,
		replicas: replicas,
		keys:     make(HashRing, 0),
		nums:     make(map[string]int),
		ring:     make(map[uint32]string),
		mu:       sync.RWMutex{},
	}
	if fn == nil {
		h.hash = crc32.ChecksumIEEE
	}
	return h
}

func (m *Map) Add(key string) {
	if _, ok := m.nums[key]; ok {
		log.Printf("%s is already exist\n", key)
		return
	}
	m.AddWithReplicas(key, m.replicas)
}

func (m *Map) AddWithReplicas(key string, replicas int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < replicas; i++ {
		hash := m.hash([]byte(strconv.Itoa(i) + key))
		m.keys = append(m.keys, hash)
		m.ring[hash] = key
		m.nums[key]++
	}
	sort.Sort(m.keys)
}

func (m *Map) Get(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.keys) == 0 || key == "" {
		return ""
	}
	hash := m.hash([]byte(key))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.ring[m.keys[idx%len(m.keys)]]
}

func (m *Map) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	counts, ok := m.nums[key]
	if !ok {
		log.Printf("%s is not exist\n", key)
		return
	}
	for i := 0; i < counts; i++ {
		hash := m.hash([]byte(strconv.Itoa(i) + key))
		for index, value := range m.keys {
			if hash == value {
				m.keys = append(m.keys[:index], m.keys[index+1:]...)
			}
		}
		delete(m.ring, hash)
		m.nums[key]--
	}
	delete(m.nums, key)
}
