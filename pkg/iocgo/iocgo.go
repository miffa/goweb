package iocgo

import (
	"iris/pkg/config"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type Initializer interface {
	Init(cfg *config.TpaasConfig) error
	Close() error
	Reload(cfg *config.TpaasConfig) error
}

var servie_pool []*ServiceItem

type ServiceItem struct {
	initptr Initializer
	name    string
}

func Register(iname string, iner Initializer) {
	item := ServiceItem{initptr: iner, name: iname}
	servie_pool = append(servie_pool, &item)
}

func LaunchEngine(cfg *config.TpaasConfig) (err error) {
	for _, initfunc := range servie_pool {
		err = initfunc.initptr.Init(cfg)
		if err != nil {
			log.Errorf("init resource[%s] err:%v", initfunc.name, err)
			return errors.Errorf("init resource[%s] err:%v", initfunc.name, err)
		}
		log.Infof("init resource[%s] ok", initfunc.name)
	}
	// when config file changed
	// 1. the config will be reloaded automatically
	// 2. notify the serverEngine and call ReloadEngineNotifier to reoload ServerEngine
	cfg.Notify(ReloadEngineNotifier)
	return nil
}

func ReloadEngineNotifier(e fsnotify.Event) {
	ReloadEngine(config.GloableCfg())
}

func ReloadEngine(cfg *config.TpaasConfig) (err error) {
	for _, initfunc := range servie_pool {
		err = initfunc.initptr.Reload(cfg)
		if err != nil {
			log.Errorf("reload resource[%s] err:%v", initfunc.name, err)
			return errors.Errorf("reload resource[%s] err:%v", initfunc.name, err)
		}
		log.Infof("reload resource[%s] ok", initfunc.name)
	}

	return nil
}

func StopEngine() (err error) {
	for _, initfunc := range servie_pool {
		err = initfunc.initptr.Close()
		if err != nil {
			log.Errorf("close resource[%s] err:%v", initfunc.name, err)
			continue
		}
		log.Infof("close resource[%s] ok", initfunc.name)
	}

	return nil
}
