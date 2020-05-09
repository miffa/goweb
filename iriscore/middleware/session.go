package middleware

import (
	"github.com/kataras/iris/v12"
)

func CheckSession(ctx iris.Context) {
	ctx.Values().Set(ReqUserKey, "xxxxxxxxx")
	ctx.Values().Set(ReqUserDomain, "xxxxxxxx")
	ctx.Next()
}
