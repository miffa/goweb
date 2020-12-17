package handler

import (
	"github.com/kataras/iris/v12"

	"iris/pkg/config"
	"iris/pkg/define"
	"iris/pkg/iocgo"
	"iris/pkg/service"

	log "github.com/sirupsen/logrus"
)

func Demo(ctx iris.Context) {

	// ctx.Params().Get("id")   // get values from restful api  //bbbbbb:8080/xxx/{id}
	// ctx.Params().GetString("id")   // get values from restful api
	// ctx.Params().GetIntXXX("id")   // get values from restful api
	// ctx.Params().GetIntXXX("id")   // get values from restful api
	// ctx.Params().GetUintXXX("id")  // get values from restful api
	//ctx.Params().GetFloatXXX("id")  // get values from restful api

	//  ctx.ReadJSON(&m)        // body is json string
	//  ctx.FormValue("name")  // get value from the URL field's query parameters and the POST or PUT form data
	//  UrlParamXXX("")        // get value from url query parameters
	//  PostValueXXX("ident")  // get value from post put and PATCH body
	retdata := service.GetSingleTon().Demook()
	ResponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}

func Demo2(ctx iris.Context) {
	retdata, err := service.GetSingleTon().Demook2()
	if err != nil {
		ResponseErrMsg(ctx, define.ST_SER_ERROR, err.Error())
		return
	}
	var user struct {
		Name string `json:"gege"`
		Age  int    `json:"age"`
	}

	mybull := ctx.Request().Header.Get("myname")
	ctx.ReadJSON(&user)
	ResponseOk(ctx, retdata)
	log.Debugf("%s demo reponse ok:%v", mybull, user)
}

func Demo3(ctx iris.Context) {
	retdata, err := service.GetSingleTon().Demook3()
	if err != nil {
		ResponseErr(ctx, err)
		return
	}
	ResponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}

func Reload(ctx iris.Context) {
	config.ReloadGloableCfg()
	iocgo.ReloadEngine(config.GloableCfg())
	ResponseOk(ctx, "OK")
}
