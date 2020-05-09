package handler

import (
	"io"
	"strconv"
	"time"
	"goweb/iriscore/config"
	"goweb/iriscore/define"
	"goweb/iriscore/iocgo"
	"goweb/iriscore/middleware/tracinglog"

	log "goweb/iriscore/libs/logrus"

	"github.com/kataras/iris/v12"
	//"github.com/kataras/iris/v12/v12/context"
	"github.com/kataras/iris/v12/context"
)

const (
	ST_OK            = 1000
	ST_ARGS_ERROR    = 401
	ST_DATA_NOTFOUND = 404
	ST_SER_ERROR     = 501
	ST_SER_BUSY      = 502
	ST_SESSION_OUT   = 601
	ST_TOKEN_OUT     = 602
	ST_AUTH_FAILURE  = 701

	ReqUserKey = "miducar_user"
	CosTimeKey = "miducar_begin_time"
)

var amr *AccessLogMgr = &AccessLogMgr{}

func init() {
	iocgo.Register("accesslog", amr)
}

// http request logger
type AccessLogMgr struct {
	AccessLog    *log.Logger
	rfAccess     io.WriteCloser
	CanISeeParam bool
}

func (am *AccessLogMgr) Init(cfg *config.TpaasConfig) error {
	am.rfAccess = log.NewRotateFile(cfg.GetString("common.access_log"), 100*log.MiB)
	am.AccessLog = log.New()
	am.AccessLog.Formatter = &log.JSONFormatter{}
	am.AccessLog.Out = am.rfAccess
	am.CanISeeParam = true
	return nil
}

func (am *AccessLogMgr) Close() error {
	am.rfAccess.Close()
	return nil
}

type HandFunc = context.Handler // func(ctx iris.Context)

type RespJson struct {
	Status int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func NewRespJson() *RespJson {
	return &RespJson{
		Status: ST_OK,
		Data:   0,
	}
}

func ResponseError(ctx iris.Context, code int, msg string, data interface{}) {
	resp := NewRespJson()
	resp.Status = code
	resp.Msg = msg
	resp.Data = data
	ctx.JSON(resp)
	responseLog(ctx, msg)
	ctx.Values().Set(define.CTX_RESP_STS_KEY, msg)
	tracinglog.FinishSpan(ctx)
	return
}

func ResponseErr(ctx iris.Context, code int, msg string) {
	resp := NewRespJson()
	resp.Status = code
	resp.Msg = msg
	ctx.JSON(resp)
	responseLog(ctx, msg)
	ctx.Values().Set(define.CTX_RESP_STS_KEY, msg)
	tracinglog.FinishSpan(ctx)
	return
}

func ReponseOk(ctx iris.Context, respdata interface{}) {
	resp := NewRespJson()
	resp.Status = ST_OK
	resp.Msg = "OK"
	resp.Data = respdata
	ctx.JSON(resp)
	ctx.Values().Set(define.CTX_RESP_STS_KEY, "OK")
	responseLog(ctx, respdata)
	tracinglog.FinishSpan(ctx)
	return
}

////////////////////////////////////////////////////////
func GetNulString(ctx iris.Context, k string) string {
	return ctx.FormValue(k)
}
func GetString(ctx iris.Context, k string) string {
	data := ctx.FormValue(k)
	if data == "" {
		ResponseErr(ctx, ST_ARGS_ERROR, k+"is empty")
		return ""
	}
	return data
}

func GetInt(ctx iris.Context, k string) int {
	data, err := strconv.Atoi(ctx.FormValue(k))
	if err != nil {
		ResponseErr(ctx, ST_ARGS_ERROR, k+"is invilid")
		return -1
	}
	return data
}
func GetInt64(ctx iris.Context, k string) int64 {
	data, err := strconv.ParseInt(ctx.FormValue(k), 10, 64)
	if err != nil {
		ResponseErr(ctx, ST_ARGS_ERROR, k+"is invalid")
		return -1
	}
	return data
}

////////////////////////////////////////////////
// access log- request
func RequestLog(ctx iris.Context) {
	var params interface{}
	params = ctx.FormValues()
	//remove params in access log
	if !amr.CanISeeParam {
		params = "request"
	}
	user := ctx.Values().Get(ReqUserKey)
	begin := time.Now()
	ctx.Values().Set(CosTimeKey, begin.UnixNano())
	//ctx.Params().Set(CosTimeKey, begin.UnixNano())

	amr.AccessLog.WithField("serino", begin.UnixNano()).WithField("path", ctx.Path()).WithField("peer", ctx.RemoteAddr()).WithField("user", user).Info(params)
	ctx.Next()
}

// access log -response
func responseLog(ctx iris.Context, respdata interface{}) {
	//middleware.FinishSpan(ctx)
	var costtime int64
	begin, err := ctx.Values().GetInt64(CosTimeKey)
	if err == nil {
		costtime = time.Now().UnixNano() - begin
		peerin := ctx.RemoteAddr()
		user := ctx.Values().Get(ReqUserKey)
		amr.AccessLog.WithField("path", ctx.Path()).
			WithField("serino", begin).
			WithField("peer", peerin).
			WithField("user", user).
			WithField("remote_ip", ctx.Request().RemoteAddr).
			WithField("costtime", costtime).
			Info(respdata)
	} else {
		costtime = 0 //time.Now().Nanosecond() - begin
		peerin := ctx.RemoteAddr()
		user := ctx.Values().Get(ReqUserKey)
		amr.AccessLog.WithField("path", ctx.Path()).
			WithField("serino", begin).
			WithField("peer", peerin).
			WithField("user", user).
			WithField("remote_ip", ctx.Request().RemoteAddr).
			WithField("costtime", costtime).
			Info(respdata)
	}
}
