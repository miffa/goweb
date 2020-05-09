package middleware

import (
	"sync"
	"time"

	"goweb/iriscore/config"
	"goweb/iriscore/handler"
	"goweb/iriscore/iocgo"

	"goweb/iriscore/libs/tokenbucket"

	iris "github.com/kataras/iris/v12"
)

const (
	VERIFICATION_USER  = "user"
	VERIFICATION_TOKEN = "token"
	AUTH_USER          = "jwtuser_logined"
	ReqUserKey         = "username_logined"
	ReqUserDomain      = "userdomain_logined"
)

type Ratelimiter struct {
	*tokenbucket.Bucket
}

var RL *Ratelimiter
var rlonce sync.Once

func Ralimit() *Ratelimiter {
	rlonce.Do(func() { RL = new(Ratelimiter) })
	return RL
}

func OnceInit() {
	RL = new(Ratelimiter)
	RL.Bucket = tokenbucket.NewBucket(1000, 1*time.Millisecond)
}

func (rl *Ratelimiter) Init(cfg *config.TpaasConfig) error {
	return rl.InitFrq(1000)
}

func (rl *Ratelimiter) InitFrq(freq int64) error {
	rl.Bucket = tokenbucket.NewBucket(freq, 1*time.Second)
	return nil
}

func (rl *Ratelimiter) Close() error {
	return rl.Bucket.Close()
}

func init() {
	iocgo.Register("ratelimit_1000", Ralimit())
}

func CheckRatelimit(ctx iris.Context) {
	if Ralimit().Take(1) != 0 {
		ctx.Next()
	} else {
		handler.ResponseErr(ctx, handler.ST_SER_BUSY, "Server is busy, ratelimited")
	}
}
