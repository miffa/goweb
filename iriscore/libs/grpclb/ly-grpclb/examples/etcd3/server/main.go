package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"goweb/iriscore/libs/grpclb/ly-grpclb/examples/proto"
	registry "goweb/iriscore/libs/grpclb/ly-grpclb/registry/etcd3"
	etcd "go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var nodeID = flag.String("node", "node1", "node ID")
var port = flag.Int("port", 8080, "listening port")
var ip = flag.String("ip", "127.0.0.1", "binding ip")

type RpcServer struct {
	addr string
	s    *grpc.Server
}

func NewRpcServer(addr string) *RpcServer {
	s := grpc.NewServer()
	rs := &RpcServer{
		addr: addr,
		s:    s,
	}
	return rs
}

func (s *RpcServer) Run() {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	log.Printf("rpc listening on:%s", s.addr)

	proto.RegisterTestServer(s.s, s)
	s.s.Serve(listener)
}

func (s *RpcServer) Stop() {
	s.s.GracefulStop()
}

func (s *RpcServer) Say(ctx context.Context, req *proto.SayReq) (*proto.SayResp, error) {
	text := "Hello " + req.Content + ", I am " + *nodeID
	log.Println(text)

	return &proto.SayResp{Content: text}, nil
}

func StartService() {
	etcdConfg := etcd.Config{
		Endpoints: []string{"http://10.104.106.89:2379"},
	}

	registry, err := registry.NewRegistry(
		registry.Option{
			EtcdConfig:  etcdConfg,
			RegistryDir: "/grpc-lb",
			ServiceName: "test",
			NodeID:      *nodeID,
			NData: registry.NodeData{
				Addr: fmt.Sprintf("%s:%d", *ip, *port),
				//Metadata: map[string]string{"weight": "1"},
			},
			Ttl: 1000 * time.Second,
		})
	if err != nil {
		log.Panic(err)
		return
	}
	server := NewRpcServer(fmt.Sprintf("0.0.0.0:%d", *port))
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		server.Run()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		registry.Register()
		wg.Done()
	}()

	//stop the server after one minute
	//go func() {
	//	time.Sleep(time.Minute)
	//	server.Stop()
	//	registry.Deregister()
	//}()

	wg.Wait()
}

//go run main.go -node node1 -port 28544
//go run main.go -node node2 -port 18562
//go run main.go -node node3 -port 27772
func main() {
	flag.Parse()
	StartService()
}
