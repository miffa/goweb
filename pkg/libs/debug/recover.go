package debug

import (
	log "github.com/sirupsen/logrus"
)

func ProtectPanic() {
	if err := recover(); err != nil {
		log.Errorf("!!!!!!!! recover from panic %v", err)
	}
}
