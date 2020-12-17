package middleware

import (
	"iris/pkg/define"
	"iris/pkg/handler"
	"iris/pkg/thirdurl/auth"
	"strings"

	"github.com/kataras/iris/v12"
)

func CheckToken(ctx iris.Context) {
	if ctx.RequestPath(false) == "/login" {
		ctx.Next()
		return
	}

	tokenstr := ctx.Request().Header.Get("Authorization")
	if tokenstr == "" {
		handler.ResponseErrMsg(ctx, define.ST_TOKEN_OUT, "token非法")
		return
	}
	tokens := strings.Split(tokenstr, " ")
	if len(tokens) != 2 {
		handler.ResponseErrMsg(ctx, define.ST_TOKEN_OUT, "token非法")
		return
	}

	if u, err := auth.TokenWithSession(tokens[1]); err != nil {
		handler.ResponseErrMsg(ctx, define.ST_TOKEN_OUT, "token非法")
		return
	} else {
		ctx.Values().Set(define.ReqUserKey, u.Name)
		ctx.Next()
	}
}
