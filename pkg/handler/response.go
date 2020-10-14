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
	//if you want to set http code,  call this function :ctx.StatusCode(your code)
	//ctx.StatusCode(resp.Code)
	ResponseLog(ctx)
	return
}

func ResponseErr(ctx iris.Context, code int, msg string) {
	resp := NewRespJson()
	resp.Code = code
	resp.Msg = msg
	ctx.Values().Set(define.CtxRespStsKey, msg)
	//if you want to set http code,  call this function :ctx.StatusCode(your code)
	//ctx.StatusCode(code)
	ctx.JSON(resp)
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
