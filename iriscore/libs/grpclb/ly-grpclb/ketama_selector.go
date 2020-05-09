package grpclb

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type KetamaSelector struct {
	baseSelector
	hash      *Ketama
	ketamaKey string
}

var (
	DefaultKetamaKey  = "grpc-lb-ketama-key"
	KetamaKeyEmptyErr = errors.New("ketama key is empty")
)

func NewKetamaSelector(ketamaKey string) Selector {
	if ketamaKey == "" {
		ketamaKey = DefaultKetamaKey
	}
	return &KetamaSelector{
		hash:         NewKetama(17, nil),
		ketamaKey:    ketamaKey,
		baseSelector: baseSelector{addrMap: make(map[string]*AddrInfo)},
	}
}

func (s *KetamaSelector) wrapAddr(addr string, idx int) string {
	return fmt.Sprintf("%s-%d", addr, idx)
}

func (s *KetamaSelector) unWrapAddr(addr string) string {
	ss := strings.Split(addr, "-")
	return ss[0]
}

func (s *KetamaSelector) Add(addr grpc.Address) error {
	fmt.Println("add", addr.Addr)
	err := s.baseSelector.Add(addr)
	if err == nil {
		a, _ := s.addrMap[addr.Addr]
		for i := 0; i < a.weight; i++ {
			s.hash.Add(s.wrapAddr(addr.Addr, i))
		}
	}
	return err
}

func (s *KetamaSelector) Delete(addr grpc.Address) error {
	a, ok := s.addrMap[addr.Addr]
	err := s.baseSelector.Delete(addr)
	if err == nil {
		if ok {
			for i := 0; i < a.weight; i++ {
				s.hash.Remove(s.wrapAddr(addr.Addr, i))
			}
		}
	}
	return err
}

func (s *KetamaSelector) Get(ctx context.Context) (addr grpc.Address, err error) {
	if len(s.addrs) == 0 {
		err = AddrListEmptyErr
		return
	}
	key, ok := ctx.Value(s.ketamaKey).(string)
	if ok {
		targetAddr, ok := s.hash.Get(key)
		if ok {
			targetAddr = s.unWrapAddr(targetAddr)
			for _, v := range s.addrs {
				if v == targetAddr {
					if addrInfo, ok := s.addrMap[v]; ok {
						if addrInfo.connected {
							addrInfo.load++
							return addrInfo.addr, nil
						}
					}
				}
			}
		} else {
			err = AddrDoesNotExistErr
		}
	} else {
		err = KetamaKeyEmptyErr
	}

	return addr, NoAvailableAddressErr
}
