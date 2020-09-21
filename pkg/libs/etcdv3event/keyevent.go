package etcdv3event

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

//  ############################
const (
	BUF_MAX_SIZE = 32 * 1024
	MSG_MAX_SIZE = 4 * 1024 * 1024

	ETCD_EVENT_ONLINE = iota
	ETCD_EVENT_OFFLINE
)

var (
	ErrID      clientv3.LeaseID = -1
	ErrNoFound error            = errors.New("key not found")
	ErrMsg     string           = ""
)

//  #########################
type EtcdNode struct {
	K   string
	V   string
	Ttl int64
}

type EtcdEvent struct {
	EtcdNode
	EventType int
}

//  ########################
type EtcdEvent3 struct {
	client3      *clientv3.Client
	Optimeout    time.Duration
	keepalivests map[string]clientv3.LeaseID
}

//  ########################
func (e *EtcdEvent3) Init(addrs []string, dialtmout, optmout time.Duration) (err error) {
	e.client3, err = clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: dialtmout,
		// etcd is implemented using grpc
		DialOptions: []grpc.DialOption{grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 60 * time.Second, Timeout: 20 * time.Second, PermitWithoutStream: true}),
			grpc.WithMaxMsgSize(MSG_MAX_SIZE),
			grpc.WithWriteBufferSize(BUF_MAX_SIZE),
			grpc.WithReadBufferSize(BUF_MAX_SIZE)},
	})

	e.keepalivests = make(map[string]clientv3.LeaseID)
	e.Optimeout = optmout
	if err != nil {
		// handle error!
		return err

	}
	return nil
}

func (e *EtcdEvent3) Shutdown() error {
	return e.client3.Close()
}

//  ################################
func (e *EtcdEvent3) SetStellarKey(data string, v string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Put(ctx, data, v)
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}

	log.Debugf("put %s:%s ok pre:%v", data, v, resp.PrevKv)
	return nil
}

func (e *EtcdEvent3) SetCometKey(data string, v string, ttl int64) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()

	ttlnode, err := e.client3.Grant(ctx, ttl)
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}

	resp, err := e.client3.Put(ctx, data, v, clientv3.WithLease(ttlnode.ID))
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}

	log.Debugf("put %s:%s ok ttl:%d pre:%v", data, v, ttl, resp.PrevKv)
	return nil
}

// put node with keepalive
func (e *EtcdEvent3) SetPlanetaryKey(data string, v string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()

	ttlnode, err := e.client3.Grant(ctx, 10)
	if err != nil {
		return err
	}

	resp, err := e.client3.Put(ctx, data, v, clientv3.WithLease(ttlnode.ID))
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}

	ch, kaerr := e.client3.KeepAlive(ctx, ttlnode.ID)
	if kaerr != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return kaerr
	}

	go func() {
		for {
			select {
			case ka, ok := <-ch:
				if !ok {
					return
				}
				log.Debugf("recv keepalive replay leaseid:%d ttl:%d", ka.ID, ka.TTL)
			}
		}
	}()

	e.keepalivests[data] = ttlnode.ID
	log.Debugf("put %s:%s ok keepalive pre:%v", data, v, resp.PrevKv)
	return nil
}

//  ########################
func (e *EtcdEvent3) ExpiredKey(id string) error {
	leaseid, ok := e.keepalivests[id]
	if !ok {
		return fmt.Errorf("not found key:%s", id)
	}
	delete(e.keepalivests, id)
	return e.expiredKey(leaseid)
}

func (e *EtcdEvent3) expiredKey(id clientv3.LeaseID) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	_, err := e.client3.Revoke(ctx, id)
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}
	log.Debugf("revoke ok(%d)", id)
	return nil
}

func (e *EtcdEvent3) DeletePrefixKey(data string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Delete(ctx, data, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}
	log.Debugf("delete ok(%d)", data, resp.Deleted)
	return nil
}

func (e *EtcdEvent3) DeleteKey(data string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Delete(ctx, data)
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}
	log.Debugf("delete ok(%d)", data, resp.Deleted)
	return nil
}

func (e *EtcdEvent3) DeleteRangeKey(data, datae string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Delete(ctx, data, clientv3.WithRange(datae))
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return err
	}
	log.Debugf("delete ok(%d)", data, resp.Deleted)
	return nil
}

//  ########################
func (e *EtcdEvent3) GetKey(data string) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Get(ctx, data)
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return ErrMsg, err
	}
	log.Debugf("get %s ok(%d)", data, resp.Count)
	if len(resp.Kvs) == 0 {
		return ErrMsg, ErrNoFound
	}
	return string(resp.Kvs[0].Value), nil
}

func (e *EtcdEvent3) GetChildKey(data string) ([]EtcdNode, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), e.Optimeout)
	defer cancelFunc()
	resp, err := e.client3.Get(ctx, data, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcd err:%s", e.etcdRespError(err))
		return nil, err
	}
	log.Debugf("get %s ok(%d)", data, len(resp.Kvs))
	if len(resp.Kvs) == 0 {
		return nil, ErrNoFound
	}
	var ret []EtcdNode
	for _, kd := range resp.Kvs {
		ed := EtcdNode{}
		ed.K = string(kd.Key)
		ed.V = string(kd.Value)
		ed.Ttl = int64(kd.Lease)
		ret = append(ret, ed)
	}
	return ret, nil
}

//  ############################
func (e *EtcdEvent3) WatchEvent(path string) <-chan *EtcdEvent {
	enentchan := make(chan *EtcdEvent, 128)
	go e.watchNodes(path, enentchan)
	return enentchan
}

func (e *EtcdEvent3) watchNodes(path string, enentchan chan *EtcdEvent) {

	rch := e.client3.Watch(context.Background(), path, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut: // add data
				log.Debugf("ETCD EVENT [%s] %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				evc := &EtcdEvent{}
				evc.K = string(ev.Kv.Key)
				evc.V = string(ev.Kv.Value)
				evc.Ttl = int64(ev.Kv.Lease)
				evc.EventType = ETCD_EVENT_ONLINE
				enentchan <- evc
			case clientv3.EventTypeDelete: // remove node
				log.Debugf("ETCD_EVENT [%s] %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				evc := &EtcdEvent{}
				evc.K = string(ev.Kv.Key)
				evc.V = string(ev.Kv.Value)
				evc.Ttl = int64(ev.Kv.Lease)
				evc.EventType = ETCD_EVENT_OFFLINE
				enentchan <- evc
			}
		}
	}
	close(enentchan)
	log.Infof("etcd watch quit(prefix key:%s)", path)
}

//  ###############################
func (e *EtcdEvent3) etcdRespError(err error) (errinfo string) {
	if err != nil {
		switch err {
		case rpctypes.ErrEmptyKey:
			errinfo = fmt.Sprintf("ErrEmptyKey")
		case rpctypes.ErrKeyNotFound:
			errinfo = fmt.Sprintf("ErrKeyNotFound")
		case rpctypes.ErrValueProvided:
			errinfo = fmt.Sprintf("ErrValueProvided")
		case rpctypes.ErrLeaseProvided:
			errinfo = fmt.Sprintf("ErrLeaseProvided")
		case rpctypes.ErrTooManyOps:
			errinfo = fmt.Sprintf("ErrTooManyOps")
		case rpctypes.ErrDuplicateKey:
			errinfo = fmt.Sprintf("ErrDuplicateKey")
		case rpctypes.ErrCompacted:
			errinfo = fmt.Sprintf("ErrCompacted")
		case rpctypes.ErrFutureRev:
			errinfo = fmt.Sprintf("ErrFutureRev")
		case rpctypes.ErrNoSpace:
			errinfo = fmt.Sprintf("ErrNoSpace")
		case rpctypes.ErrLeaseNotFound:
			errinfo = fmt.Sprintf("ErrLeaseNotFound")
		case rpctypes.ErrLeaseExist:
			errinfo = fmt.Sprintf("ErrLeaseExist")
		case rpctypes.ErrLeaseTTLTooLarge:
			errinfo = fmt.Sprintf("ErrLeaseTTLTooLarge")
		case rpctypes.ErrMemberExist:
			errinfo = fmt.Sprintf("ErrMemberExist")
		case rpctypes.ErrPeerURLExist:
			errinfo = fmt.Sprintf("ErrPeerURLExist")
		case rpctypes.ErrMemberNotEnoughStarted:
			errinfo = fmt.Sprintf("ErrMemberNotEnoughStarted")
		case rpctypes.ErrMemberBadURLs:
			errinfo = fmt.Sprintf("ErrMemberBadURLs")
		case rpctypes.ErrMemberNotFound:
			errinfo = fmt.Sprintf("ErrMemberNotFound")
		case rpctypes.ErrRequestTooLarge:
			errinfo = fmt.Sprintf("ErrRequestTooLarge")
		case rpctypes.ErrTooManyRequests:
			errinfo = fmt.Sprintf("ErrTooManyRequests")
		case rpctypes.ErrRootUserNotExist:
			errinfo = fmt.Sprintf("ErrRootUserNotExist")
		case rpctypes.ErrRootRoleNotExist:
			errinfo = fmt.Sprintf("ErrRootRoleNotExist")
		case rpctypes.ErrUserAlreadyExist:
			errinfo = fmt.Sprintf("ErrUserAlreadyExist")
		case rpctypes.ErrUserEmpty:
			errinfo = fmt.Sprintf("ErrUserEmpty")
		case rpctypes.ErrUserNotFound:
			errinfo = fmt.Sprintf("ErrUserNotFound")
		case rpctypes.ErrRoleAlreadyExist:
			errinfo = fmt.Sprintf("ErrRoleAlreadyExist")
		case rpctypes.ErrRoleNotFound:
			errinfo = fmt.Sprintf("ErrRoleNotFound")
		case rpctypes.ErrAuthFailed:
			errinfo = fmt.Sprintf("ErrAuthFailed")
		case rpctypes.ErrPermissionDenied:
			errinfo = fmt.Sprintf("ErrPermissionDenied")
		case rpctypes.ErrRoleNotGranted:
			errinfo = fmt.Sprintf("ErrRoleNotGranted")
		case rpctypes.ErrPermissionNotGranted:
			errinfo = fmt.Sprintf("ErrPermissionNotGranted")
		case rpctypes.ErrAuthNotEnabled:
			errinfo = fmt.Sprintf("ErrAuthNotEnabled")
		case rpctypes.ErrInvalidAuthToken:
			errinfo = fmt.Sprintf("ErrInvalidAuthToken")
		case rpctypes.ErrInvalidAuthMgmt:
			errinfo = fmt.Sprintf("ErrInvalidAuthMgmt")
		case rpctypes.ErrNoLeader:
			errinfo = fmt.Sprintf("ErrNoLeader")
		case rpctypes.ErrNotLeader:
			errinfo = fmt.Sprintf("ErrNotLeader")
		case rpctypes.ErrLeaderChanged:
			errinfo = fmt.Sprintf("ErrLeaderChanged")
		case rpctypes.ErrNotCapable:
			errinfo = fmt.Sprintf("ErrNotCapable")
		case rpctypes.ErrStopped:
			errinfo = fmt.Sprintf("ErrStopped")
		case rpctypes.ErrTimeout:
			errinfo = fmt.Sprintf("ErrTimeout")
		case rpctypes.ErrTimeoutDueToLeaderFail:
			errinfo = fmt.Sprintf("ErrTimeoutDueToLeaderFail")
		case rpctypes.ErrTimeoutDueToConnectionLost:
			errinfo = fmt.Sprintf("ErrTimeoutDueToConnectionLost")
		case rpctypes.ErrUnhealthy:
			errinfo = fmt.Sprintf("ErrUnhealthy")
		case rpctypes.ErrCorrupt:
			errinfo = fmt.Sprintf("ErrCorrupt")
		case context.Canceled:
			errinfo = fmt.Sprintf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			errinfo = fmt.Sprintf("ctx is attached with a deadline is exceeded: %v", err)
		default:
			errinfo = fmt.Sprintf("what th error is: %v", err)
		}
	}
	return
}
