package gemcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/iam1912/gemcache/consistenthash"
	pb "github.com/iam1912/gemcache/proto"
	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath = "/_gemcache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

var _ PeerPicker = (*HTTPPool)(nil)

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no groupname "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := proto.Marshal(&pb.GetResponse{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

type httpGetter struct {
	baseURL string
}

var _ PeerGetter = (*httpGetter)(nil)

func (h *httpGetter) Get(in *pb.GetRequest, out *pb.GetResponse) error {
	path := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	resp, err := http.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
