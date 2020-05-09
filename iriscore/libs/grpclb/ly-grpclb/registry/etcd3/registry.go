package etcd

import (
	"encoding/json"
	"time"

	etcd3 "go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/grpclog"
)

type EtcdReigistry struct {
	etcd3Client *etcd3.Client
	key         string
	value       string
	ttl         time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	respID      etcd3.LeaseID
}

type Option struct {
	EtcdConfig  etcd3.Config
	RegistryDir string
	ServiceName string
	NodeID      string
	NData       NodeData
	Ttl         time.Duration
}

type NodeData struct {
	Addr     string
	Metadata map[string]string
}

func NewRegistry(option Option) (*EtcdReigistry, error) {
	client, err := etcd3.New(option.EtcdConfig)
	if err != nil {
		return nil, err
	}

	val, err := json.Marshal(option.NData)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	registry := &EtcdReigistry{
		etcd3Client: client,
		key:         option.RegistryDir + "/" + option.ServiceName + "/" + option.NodeID,
		value:       string(val),
		ttl:         option.Ttl,
		ctx:         ctx,
		cancel:      cancel,
	}
	return registry, nil
}

func (e *EtcdReigistry) Register() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	resp, err := e.etcd3Client.Grant(ctx, int64(e.ttl))
	if err != nil {
		return err
	}
	if _, err := e.etcd3Client.Put(e.ctx, e.key, e.value, etcd3.WithLease(resp.ID)); err != nil {
		grpclog.Printf("grpclb: set key '%s' with ttl to etcd3 failed: %s", e.key, err.Error())
		return err
	}

	ch, kaerr := e.etcd3Client.KeepAlive(ctx, resp.ID)
	if kaerr != nil {
		grpclog.Printf("etcd err:%s", err.Error())
		return kaerr
	}
	e.respID = resp.ID

	go func() {
		for {
			select {
			case ka, ok := <-ch:
				if !ok {
					return
				}
				grpclog.Printf("recv keepalive replay leaseid:%d ttl:%d", ka.ID, ka.TTL)
			}
		}
	}()

	return nil
}

func (e *EtcdReigistry) Deregister() error {
	_, err := e.etcd3Client.Revoke(context.Background(), e.respID)
	return err
}
