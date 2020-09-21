package middleware

import (
	"github.com/sirupsen/logrus"

	iris "github.com/kataras/iris/v12"
)

func DemoMiddleware(ctx iris.Context) {
	logrus.Infof("in demo middleware")
	ctx.Next()
}
func DemoPartyMiddleware(ctx iris.Context) {
	logrus.Infof("in party middleware")
	ctx.Next()
}
