package etcdevent

import (
	"fmt"
	"strconv"
	"time"

	log "goweb/iriscore/libs/logrus"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var (
	ring              *RingBell
	SER_KeepAlive     = 3 * time.Minute
	KeepAliveInterval = 2 * time.Minute
)

///////////////////////////////
type RingBell struct {
	KeysAPI  client.KeysAPI
	Etcdpath string
}

func InitRingBell(endpoints []string) error {

	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatalf("Error: cannot connec to etcd:%v", err)
		return err
	}

	ring = &RingBell{
		KeysAPI: client.NewKeysAPI(etcdClient),
	}

	return nil
}

func (m *RingBell) Init(endpoints []string) error {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		return err
	}

	m.KeysAPI = client.NewKeysAPI(etcdClient)

	return nil

}

func (m *RingBell) WatchNotifyOnce(path string, refresh chan struct{}) {
	api := m.KeysAPI
	watcher := api.Watcher(path, &client.WatcherOptions{})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			log.Errorf("Error watch workers:%v", err)
			break
		}
		//Action include get, set, delete, update, create, compareAndSwap,
		// compareAndDelete and expire
		switch res.Action {
		case "expire":
			fallthrough
		case "set":
			fallthrough
		case "delete":
			fallthrough
		case "update":
			fallthrough
		case "create":
			refresh <- struct{}{}
			log.Infof("receive notify, task maybe ok")
		}
		return
	}
}

func (m *RingBell) WatchNotify(path string, refresh chan struct{}) {
	api := m.KeysAPI
	watcher := api.Watcher(path, &client.WatcherOptions{
		Recursive: true,
	})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			log.Errorf("Error watch workers:%v", err)
			break
		}
		//Action include get, set, delete, update, create, compareAndSwap,
		// compareAndDelete and expire
		switch res.Action {
		case "expire":
			fallthrough
		case "set":
			fallthrough
		case "delete":
			fallthrough
		case "update":
			fallthrough
		case "create":
			refresh <- struct{}{}
			log.Infof("receive notify, task maybe ok")
		}
	}
	log.Infof("etcd watch routine is quiting")
}

func (m *RingBell) Notify(path string) {
	api := m.KeysAPI

	key := path + "YOU_NEED_FRESH"
	notifytime := time.Now()

	_, err := api.Set(context.Background(), key, notifytime.String(), &client.SetOptions{})
	if err != nil {
		log.Errorf("Error update workerInfo:%v", err)
	}
}

func (m *RingBell) Register(key, data string, ttl time.Duration) error {

	sopt := client.SetOptions{}
	sopt.TTL = ttl
	_, err := m.KeysAPI.Set(context.Background(), key, data, &sopt)
	if err != nil {
		log.Errorf("Error register %s:%s:%v", key, data, err)
	}
	log.Infof("Set %s:%s ok ", key, data)
	return nil
}

func (m *RingBell) KeepAlive(name, data, path string) error {
	kt := time.NewTimer(KeepAliveInterval)
	for {
		select {
		case <-kt.C:
			if err := m.Register(path+name, data, SER_KeepAlive); err != nil {
				log.Errorf("service keep alive register %s:%s err:%v", name, data, err)
			}
			kt.Reset(KeepAliveInterval)
		}
	}
	return nil
}

func (m *RingBell) GetData(path string) ([]string, error) {
	resp, err := m.KeysAPI.Get(context.Background(), path, &client.GetOptions{})
	if err != nil {
		return nil, err
	}

	if !resp.Node.Dir {
		return nil, fmt.Errorf("path is not dir in etcd")
	}
	var nodes []string
	for _, node := range resp.Node.Nodes {
		//log.Infof("%s:%s", node.Key, node.Value)
		nodes = append(nodes, node.Value)
	}
	return nodes, nil

}

func (m *RingBell) SetWithTtl(key, data string, ttl time.Duration) error {

	sopt := client.SetOptions{}
	sopt.TTL = ttl
	_, err := m.KeysAPI.Set(context.Background(), key, data, &sopt)
	if err != nil {
		log.Errorf("Error register %s:%s:%v", key, data, err)
	}
	log.Debugf("Set %s:%s ok ", key, data)
	return nil
}

func (m *RingBell) Set(key, data string) error {

	sopt := client.SetOptions{}
	_, err := m.KeysAPI.Set(context.Background(), key, data, &sopt)
	if err != nil {
		log.Errorf("Error register %s:%s:%v", key, data, err)
	}
	log.Debugf("Set %s:%s ok ", key, data)
	return nil
}

func keyNotFound(err error) bool {
	if err != nil {
		if etcdError, ok := err.(client.Error); ok {
			if etcdError.Code == client.ErrorCodeKeyNotFound ||
				etcdError.Code == client.ErrorCodeNotFile ||
				etcdError.Code == client.ErrorCodeNotDir {
				return true
			}
		}
	}
	return false
}

func (m *RingBell) GetNewId(path, name string) (string, error) {

	key := path + name
	sopt := client.SetOptions{}
	ts1 := false
	data := "10000"

	resp, err := m.KeysAPI.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		if keyNotFound(err) {
			sopt.PrevExist = client.PrevNoExist
			ts1 = true
		} else {
			return "", err
		}
	}
	//log.Debugf("%s:%s:%s", resp.Node.Key, resp.Node.Value, resp.Node.String())

	if !ts1 {
		sopt.PrevValue = resp.Node.Value
		sopt.PrevIndex = resp.Node.ModifiedIndex
		var index int
		index, err = strconv.Atoi(resp.Node.Value)
		if err != nil {
			return "", err
		}
		data = strconv.Itoa(index + 1)
	}

	_, err = m.KeysAPI.Set(context.Background(), key, data, &sopt)
	if err != nil {
		if etcdError, ok := err.(client.Error); ok {
			// Compare failed
			if etcdError.Code == client.ErrorCodeTestFailed {
				log.Warnf("set  %s:%s :Unable to complete atomic operation, key modified", key, data)
				return "", nil
			}
			// Node exists error (when PrevNoExist)
			if etcdError.Code == client.ErrorCodeNodeExist {
				log.Warnf(" %s:%s:Previous K/V pair exists, cannot complete Atomic operation", key, data)
				return "", nil
			}
		}
		log.Errorf("Error automic set  %s:%s:%v", key, data, err)
		return "", err
	}
	log.Debugf("get new index %s:%s ok ", key, data)
	return data, nil
}

func (m *RingBell) GetValue(path string) (string, error) {
	resp, err := m.KeysAPI.Get(context.Background(), path, &client.GetOptions{})
	if err != nil {
		return "", err
	}

	if resp.Node.Dir {
		return "", fmt.Errorf("path is dir in etcd, not a key")
	}
	return resp.Node.Value, nil
}

func (m *RingBell) GetDataMap(path string) (map[string]string, error) {
	resp, err := m.KeysAPI.Get(context.Background(), path, &client.GetOptions{})
	if err != nil {
		return nil, err
	}

	if !resp.Node.Dir {
		return nil, fmt.Errorf("path is not dir in etcd")
	}
	ags := make(map[string]string)
	for _, node := range resp.Node.Nodes {
		//log.Infof("%s:%s", node.Key, node.Value)
		ags[node.Key] = node.Value
	}
	return ags, nil
}

func (m *RingBell) RaceToControl(key, value string) error {
	setOptions := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       5 * time.Second,
	}
	resp, err := m.KeysAPI.Set(context.Background(), key, value, setOptions)
	if err == nil {
		log.Debugf("Create node %v OK [%q]", key, resp)
		return nil
	}
	e, ok := err.(client.Error)
	if !ok {
		return err
	}

	if e.Code != client.ErrorCodeNodeExist {
		return err
	}
	return nil
}

func (m *RingBell) ReleaseControl(key string) (err error) {
	resp, err := m.KeysAPI.Delete(context.Background(), key, nil)
	if err == nil {
		log.Debugf("Delete %s %q OK", key, resp)
		return nil
	}
	e, ok := err.(client.Error)
	if ok && e.Code == client.ErrorCodeKeyNotFound {
		return nil
	}
	log.Debugf("Delete %v falied: %q", key, resp)
	return err
}
