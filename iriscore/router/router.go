package router

import (
	"goweb/iriscore/handler"
	"goweb/iriscore/middleware"
	"goweb/iriscore/middleware/tracinglog"
)

////////////////////////////////
// router
func (a *API) InitRouter() *API {
	//init middleware here
	a.SetMiddleware(middleware.IAmAlive)
	a.SetMiddleware(handler.RequestLog)
	//a.SetMiddleware(middleware.DemoMiddleware)
	a.SetMiddleware(middleware.CheckRatelimit)
	a.SetMiddleware(tracinglog.Tracing)
	a.SetDone(tracinglog.FinishSpan)

	// keepalived api
	a.Get("/do_not_delete.html", nil)

	//global api demo
	{
		a.Get("/demoget", handler.Demo)
		a.Post("/demopost", handler.Demo2)
	}

	// group routing api demo
	{
		p := a.Group("/swordmen_novel", middleware.DemoPartyMiddleware)
		//p.Use(middleware.DemoPartymiddleware)
		p.Get("/gulong", handler.Demo)    //api: http://xxxxx:xxxx/swordmen_novel/gulong
		p.Post("/jinyong", handler.Demo2) //api: http://xxxxx:xxxx/swordmen_novel/jinyong
	}

	return a
}
