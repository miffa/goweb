package handler

import (
	"github.com/kataras/iris/v12"

	"goweb/iriscore/service"

	log "goweb/iriscore/libs/logrus"
)

func Demo(ctx iris.Context) {
	retdata := service.GetSingleTon().Demook()
	ReponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}

func Demo2(ctx iris.Context) {
	retdata := service.GetSingleTon().Demook2()
	ReponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}
