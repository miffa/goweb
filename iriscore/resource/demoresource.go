package resource

import (
	"fmt"
	"sync"

	"goweb/iriscore/config"
	"goweb/iriscore/iocgo"
)

var demosrc *Demoresorce
var once sync.Once

func SigleTon() *Demoresorce {
	once.Do(func() { demosrc = &Demoresorce{} })
	return demosrc
}

///////////// resource skeleton///////////////

type Demoresorce struct {
	democlient *Xclient
}

func (d *Demoresorce) Init(cfg *config.TpaasConfig) error {
	d.democlient = &Xclient{Addr: cfg.GetString("mysql.datasource")}
	return d.democlient.Connect()
}

func (d *Demoresorce) Close() error {
	return d.democlient.Stop()
}

func (d *Demoresorce) YourFunction(args interface{}) string {
	//skeleton
	return fmt.Sprintf("%v", args)
}

func init() {
	iocgo.Register("Demoresorce pool", SigleTon())
}

//////////// connection resource demo , ignore //////////
type Xclient struct {
	//some resorce
	Addr string
}

func (x *Xclient) Connect() error {
	//some client connection
	// todo:
	return nil
}

func (x *Xclient) Stop() error {
	// close service resource
	return nil
}
