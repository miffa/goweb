package middleware

import (
	iris "github.com/kataras/iris/v12"
)

// IAmAlive: check alive:w
func IAmAlive(ctx iris.Context) {
	if ctx.RequestPath(false) == "/do_not_delete.html" {
		ctx.Text("I am alive. Please feel free to use it")
	} else {
		ctx.Next()
	}
}
