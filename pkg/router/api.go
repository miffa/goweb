package router

import (
	"io"
	"sync"
	"time"

	"iris/pkg/handler"

	stdContext "context"

	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/core/router"
	"github.com/valyala/tcplisten"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
)

var mainapi *API
var onemoon sync.Once
var Idletimeout time.Duration = 20 * time.Second
var Readtimeout time.Duration = 20 * time.Second
var Writetimeout time.Duration = 20 * time.Second

type API struct {
	*iris.Application
	grouprouter map[string]*ChildRouter
}

type ChildRouter struct {
	router.Party
	grouprouter map[string]*ChildRouter
}

///////
// singleton api object
func Api() *API {
	onemoon.Do(
		func() {
			mainapi = &API{}
			mainapi.Application = iris.New()
			mainapi.Use(recover.New())
			mainapi.grouprouter = make(map[string]*ChildRouter)
		})
	return mainapi
}

func (a *API) Shutdown() {
	a.Application.Shutdown(stdContext.Background())
}

func (a *API) ConfigDefault() *API {
	return a
}

func (a *API) SetTimeout(d time.Duration) *API {
	Idletimeout = d
	Readtimeout = d
	Writetimeout = d
	a.Application.ConfigureHost(Timeout)
	return a
}

func Timeout(su *host.Supervisor) {
	su.Server.IdleTimeout = Idletimeout
	su.Server.ReadTimeout = Readtimeout
	su.Server.WriteTimeout = Writetimeout
}

//// set logger
func (a *API) SetLog(w io.Writer) *API {
	a.Logger().SetOutput(w)
	return a
}

func (a *API) Runapi(ippost string) error {
	listenerCfg := tcplisten.Config{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    false,
	}

	l, err := listenerCfg.NewListener("tcp4", ippost)
	if err != nil {
		return err
	}

	//go a.Run(iris.Listener(l))
	go a.Run(iris.Listener(l))
	return nil
}

func (a *API) Websocket() {
}

//////////////////////////////
// set midware
func (a *API) SetMiddleware(m handler.HandFunc) *API {
	a.Use(m)
	return a
}

// set defer func for http request
func (a *API) SetDone(m handler.HandFunc) *API {
	a.Done(m)
	return a
}

///// group routing

func (a *API) Group(p string, f ...handler.HandFunc) *ChildRouter {
	groupiter := a.Party(p, f...)
	cr := &ChildRouter{Party: groupiter, grouprouter: make(map[string]*ChildRouter)}
	a.grouprouter[p] = cr
	return cr
}

func (c *ChildRouter) Group(p string, f ...handler.HandFunc) *ChildRouter {
	groupiter := c.Party.Party(p, f...)
	cr := &ChildRouter{Party: groupiter, grouprouter: make(map[string]*ChildRouter)}
	c.grouprouter[p] = cr
	return cr
}
