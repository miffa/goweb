package etcdv3event

import (
	"context"
	"errors"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

//  ############################
const (
	US_P_INIT = iota
	US_P_RUNNING
	US_P_ELECT
	US_P_QUIT
	US_P_ERROR
)

var (
	//ErrID      clientv3.LeaseID = -1
	//ErrNoFound error            = errors.New("key not found")
	//ErrMsg     string           = ""
	ErrStat error = errors.New("Your stat is not correct")
)

//  ########################
type EtcdYouToBeNo1 struct {
	client3      *clientv3.Client
	Optimeout    time.Duration
	leader       *concurrency.Election
	session      *concurrency.Session
	uuid         string
	cancelFunc   context.CancelFunc
	campaignStat int32
}

func (e *EtcdYouToBeNo1) Init(addrs []string, dialtmout, optmout time.Duration, servicename string) (err error) {
	e.client3, err = clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: dialtmout,
		DialOptions: []grpc.DialOption{grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 60 * time.Second, Timeout: 20 * time.Second, PermitWithoutStream: true}),
			grpc.WithMaxMsgSize(MSG_MAX_SIZE),
			grpc.WithWriteBufferSize(BUF_MAX_SIZE),
			grpc.WithReadBufferSize(BUF_MAX_SIZE)},
	})

	e.Optimeout = optmout
	if err != nil {
		// handle error!
		return err

	}

	e.session, err = concurrency.NewSession(e.client3)
	if err != nil {
		return err

	}

	e.leader = concurrency.NewElection(e.session, servicename)
	uid, _ := uuid.NewV4()
	e.uuid = strings.Trim(servicename, "/") + "_" + uid.String()
	e.campaignStat = US_P_INIT
	return nil
}

// block until to be a leader
func (e *EtcdYouToBeNo1) RunForPresident() error {
	if e.campaignStat != US_P_INIT {
		return ErrStat
	}
	e.campaignStat = US_P_RUNNING
	cctx, cancel := context.WithCancel(context.TODO())
	e.cancelFunc = cancel
	if err := e.leader.Campaign(cctx, e.uuid); err != nil {
		log.Error("RunForPresident err:%v", err)
		e.campaignStat = US_P_ERROR
		return err
	}
	e.campaignStat = US_P_ELECT
	return nil
}

// Resign lets a leader start a new election. (delete somethong in etcd)
func (e *EtcdYouToBeNo1) Resign() error {
	if e.campaignStat != US_P_ELECT {
		return nil
	}
	e.campaignStat = US_P_QUIT
	if err := e.leader.Resign(context.TODO()); err != nil {
		e.campaignStat = US_P_ERROR
		return err
	}
	e.campaignStat = US_P_INIT
	return nil
}

// EndOfTerm let leader
func (e *EtcdYouToBeNo1) EndOfTerm() error {
	if e.campaignStat == US_P_ELECT {
		e.campaignStat = US_P_QUIT
		e.Resign()
	} else if e.campaignStat == US_P_RUNNING {
		e.campaignStat = US_P_QUIT
		if e.cancelFunc != nil {
			e.cancelFunc()
		}
	}
	e.session.Close()
	e.client3.Close()
	return nil
}

//CompareAndSwapInt32(ptr, oldint, newint)
//AddInt32(ptr, int32)
//LoadInt32(ptr)
//StoreInt32(ptr, int32)

//  ###################
// demo

/*
    el := &EtcdYouToBeNo1{}
	el.Init(.....)
	e.RunForPresident()

	defer el.EndOfTerm()
    //todo :
	something that president do
*/
