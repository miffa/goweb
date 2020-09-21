package handler

import (
	"iris/pkg/define"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kataras/iris/v12"
	//"github.com/kataras/iris/v12/v12/context"
	"github.com/kataras/iris/v12/context"
)

const (
	ResponseStatusOk = "OK"
)

type HandFunc = context.Handler // func(ctx iris.Context)

type RespJson struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Detail string      `json:detail, omitempty`
}

func NewRespJson() *RespJson {
	return &RespJson{
		Code:   define.ST_OK,
		Data:   0,
		Detail: "",
	}
}

////////////////////////////////////////////////////////
func GetNulString(ctx iris.Context, k string) string {
	return ctx.FormValue(k)
}
func GetString(ctx iris.Context, k string) string {

	data := ctx.FormValue(k)
	if data == "" {
		ResponseErr(ctx, define.ST_ARGS_ERROR, k+"is empty")
		return ""
	}
	return data
}

func GetInt(ctx iris.Context, k string) int {
	data, err := strconv.Atoi(ctx.FormValue(k))
	if err != nil {
		ResponseErr(ctx, define.ST_ARGS_ERROR, k+"is invilid")
		return -1
	}
	return data
}
func GetInt64(ctx iris.Context, k string) int64 {
	data, err := strconv.ParseInt(ctx.FormValue(k), 10, 64)
	if err != nil {
		ResponseErr(ctx, define.ST_ARGS_ERROR, k+"is invalid")
		return -1
	}
	return data
}

////////////////////////////////////////////////
// access log- request
func RequestLog(ctx iris.Context) {
	var params interface{}
	params = ctx.FormValues()
	user := ctx.Values().Get(define.ReqUserKey)
	begin := time.Now()
	ctx.Values().Set(define.CosTimeKey, begin.UnixNano())

	log.WithField("serino", begin.UnixNano()).WithField("path", ctx.Path()).WithField("peer", ctx.RemoteAddr()).WithField("user", user).Info(params)
	ctx.Next()
}

// access log -response
func ResponseLog(ctx iris.Context) {
	var costtime int64
	begin, _ := ctx.Values().GetInt64(define.CosTimeKey)

	costtime = time.Now().UnixNano() - begin
	peerin := ctx.RemoteAddr()
	user := ctx.Values().Get(define.ReqUserKey)
	respdata := ctx.Values().Get(define.CtxRespStsKey).(string)

	if ctx.Method() == http.MethodGet {
		log.WithField("path", ctx.Path()).
			WithField("serino", begin).
			WithField("peer", peerin).
			WithField("user", user).
			WithField("remote_ip", ctx.Request().RemoteAddr).
			WithField("costtime", costtime).Debug(respdata)

	} else {
		log.WithField("path", ctx.Path()).
			WithField("serino", begin).
			WithField("peer", peerin).
			WithField("user", user).
			WithField("remote_ip", ctx.Request().RemoteAddr).
			WithField("costtime", costtime).Info(respdata)
	}
}
