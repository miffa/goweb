package handler

import (
	"iris/pkg/define"

	"github.com/kataras/iris/v12"
)

func ResponseErrMsg(ctx iris.Context, err error) {
	resp := NewRespJson()
	resp.Code, resp.Msg, resp.Detail = define.ErrorMsg(err)
	ctx.JSON(resp)
	ctx.Values().Set(define.CtxRespStsKey, resp.Detail)
	//ctx.Next() if api  err, no need process Done middleware
	ResponseLog(ctx)
	return
}

func ResponseErr(ctx iris.Context, code int, msg string) {
	resp := NewRespJson()
	resp.Code = code
	resp.Msg = msg
	ctx.Values().Set(define.CtxRespStsKey, msg)
	ctx.JSON(resp)
	//ctx.Next()   if api  err, no need process Done middleware
	ResponseLog(ctx)
	return
}

func ResponseOk(ctx iris.Context, respdata interface{}) {
	resp := NewRespJson()
	resp.Code = define.ST_OK
	resp.Msg = ResponseStatusOk
	resp.Data = respdata
	ctx.Values().Set(define.CtxRespStsKey, ResponseStatusOk)
	ctx.JSON(resp)
	ctx.Next()
	return
}
