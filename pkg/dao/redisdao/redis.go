package redisdao

import (
	"time"

	"github.com/garyburd/redigo/redis"

	log "github.com/sirupsen/logrus"
)

type RedisMgr struct {
	redisOp *redis.Pool
}

func (r *RedisMgr) Init(addr string, maxidle, maxactive int, timeout time.Duration) error {
	r.redisOp = &redis.Pool{
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: timeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				log.Errorf("redis.Dial(tcp, %s) err:%v", addr, err)
				return nil, err
			}
			return c, nil
		},
	}
	conn := r.redisOp.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return err
	}

	return nil
}

func (r *RedisMgr) Close() {
	r.redisOp.Close()
}

func (r *RedisMgr) Get() redis.Conn {
	return r.redisOp.Get()
}

// put ;  call  redis.Conn.Close() putting conn to the pool
