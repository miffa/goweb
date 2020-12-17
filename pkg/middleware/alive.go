package middleware

import (
	iris "github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
)

// IAmAlive: check alive:w
func IAmAlive(ctx iris.Context) {
	if ctx.RequestPath(false) == "/do_not_delete.html" {
		ctx.Text("I am alive. Please feel free to use it")
	} else {
		ctx.Next()
	}
}

func Log404Error(ctx iris.Context) {
	logError(404, ctx)
}

func Log403Error(ctx iris.Context) {
	logError(403, ctx)
}

func Log500Error(ctx iris.Context) {
	logError(500, ctx)
}

func Log502Error(ctx iris.Context) {
	logError(502, ctx)
}

func logError(code int, ctx iris.Context) {
	logrus.Errorf("url:%s response is %d", ctx.Request().URL.String(), code)
	ctx.Writef("url:%s response is %d", ctx.Request().URL.String(), code)
}
