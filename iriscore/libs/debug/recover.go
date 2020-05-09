package debug

import (
	log "goweb/iriscore/libs/logrus"
)

func ProtectPanic() {
	if err := recover(); err != nil {
		log.Errorf("!!!!!!!! recover from panic %v", err)
	}
}
