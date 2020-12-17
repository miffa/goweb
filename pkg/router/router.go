package router

import (
	"iris/pkg/handler"
	"iris/pkg/middleware"
	"net/http"
)

////////////////////////////////
// router
func (a *API) InitRouter() *API {
	//init middleware here

	a.OnErrorCodeLog(http.StatusNotFound, middleware.Log404Error)
	a.OnErrorCodeLog(http.StatusInternalServerError, middleware.Log500Error)
	a.OnErrorCodeLog(http.StatusBadGateway, middleware.Log502Error)
	a.SetMiddleware(middleware.IAmAlive)
	a.SetMiddleware(handler.RequestLog)

	// check session
	//a.SetMiddleware(middleware.CheckToken)

	// keepalived api
	a.Get("/do_not_delete.html", nil)
	a.Any("/-/reload", handler.Reload)

	//global api demo
	{
		a.Get("/demoget", handler.Demo)
		a.Get("/demoget3", handler.Demo3)
		a.Post("/demopost", handler.Demo2)
	}

	// group routing api demo
	{
		p := a.Group("/api/v1/swordmen_novel", middleware.DemoPartyMiddleware)
		//p.Use(middleware.DemoPartymiddleware)
		p.Get("/gulong", handler.Demo)    //api: http://xxxxx:xxxx/swordmen_novel/gulong
		p.Post("/jinyong", handler.Demo2) //api: http://xxxxx:xxxx/swordmen_novel/jinyong
	}

	return a
}
