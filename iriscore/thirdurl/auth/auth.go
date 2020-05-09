package auth

import (
	"goweb/iriscore/config"
	"goweb/iriscore/iocgo"

	iris "github.com/kataras/iris/v12"
)

var (
	UAUTH_URL = "" //uauth domain
)

type UauthSource struct {
}

func (d *UauthSource) Init(cfg *config.TpaasConfig) error {
	InitUauthCon()
	return nil
}

func (d *UauthSource) Close() error {
	return nil
}

func init() {
	iocgo.Register("uauth service", new(UauthSource))
}

func InitUauthCon() {
}

func CanIgo(ctx iris.Context, resource, action string) (bool, error) {
	return false, nil
}
