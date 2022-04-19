package gemcache

import pb "github.com/iam1912/gemcache/proto"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(req *pb.GetRequest, resp *pb.GetResponse) error
}
