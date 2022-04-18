package gemcache

import (
	"context"
	"fmt"
	"net"

	pb "github.com/iam1912/gemcache/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GemCacheServe struct {
	self        string
	baseURLPath string
}

func NewGemCacheServe(self, baseURLPath string) *GemCacheServe {
	return &GemCacheServe{
		self:        self,
		baseURLPath: baseURLPath,
	}
}

func (s *GemCacheServe) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	group := gropus[req.Group]
	if group == nil {
		return nil, fmt.Errorf("no groupname: %s", req.Group)
	}
	value, err := group.Get(req.Key)
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Value: value.ByteSlice()}, nil
}

func HandleService(self, baseURLPath string) {
	s := grpc.NewServer()
	gemServe := NewGemCacheServe(self, baseURLPath)
	pb.RegisterGroupCacheServer(s, gemServe)
	l, err := net.Listen("tcp", self)
	if err != nil {
		panic(err)
	}
	s.Serve(l)
}

func GemcacheClientReq(addr string, group, key string) ([]byte, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewGroupCacheClient(conn)
	resp, err := c.Get(context.Background(), &pb.GetRequest{Group: group, Key: key})
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}
