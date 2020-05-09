package main

import (
	"log"
	"time"

	grpclb "goweb/iriscore/libs/grpclb/ly-grpclb"
	"goweb/iriscore/libs/grpclb/ly-grpclb/examples/proto"
	registry "goweb/iriscore/libs/grpclb/ly-grpclb/registry/etcd3"
	etcd "go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	etcdConfg := etcd.Config{
		Endpoints: []string{"http://10.104.106.89:2379"},
	}
	r := registry.NewResolver("/grpc-lb", "test", etcdConfg)
	b := grpclb.NewBalancer(r, grpclb.NewKetamaSelector("yml"))
	//b = grpclb.NewBalancer(r, grpclb.NewRoundRobinSelector())
	c, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithBalancer(b), grpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Printf("grpc dial: %s", err)
		return
	}
	defer c.Close()

	client := proto.NewTestClient(c)

	time.Sleep(5 * time.Second)

	for i := 0; i < 50; i++ {
		log.Println("aa===============")
		ctx := context.Background()
		switch i % 5 {
		case 0:
			ctx = context.WithValue(context.Background(), "yml", "gagafdsgfdgfdgfdg")
		case 1:
			ctx = context.WithValue(context.Background(), "yml", "zzzzzzzzzzzzz")
		case 2:
			ctx = context.WithValue(context.Background(), "yml", "#$%^&((*&^&))")
		case 3:
			ctx = context.WithValue(context.Background(), "yml", "43yy567i75")
		case 4:
			ctx = context.WithValue(context.Background(), "yml", "gfhmhntwj")
		}
		ctxx, funx := context.WithTimeout(ctx, 5*time.Second)
		defer funx()
		resp, err := client.Say(ctxx, &proto.SayReq{Content: "round robin"})
		if err != nil {
			log.Println("aa:", err)
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)
		log.Printf(resp.Content)
	}

}
