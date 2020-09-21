package tpaashttpclient

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func UrlString(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func HttpOK(i int) bool {
	return i >= 200 && i < 226 //  from https://golang.org/src/net/http/status.go
}

func Http200(i int) bool {
	return i == 200
}

func HttpErr(i int) bool {
	return i >= 400 && i <= 511
}

func Begin() time.Time {
	return time.Now()
}

func End(ctl string, beg time.Time) {
	logrus.Infof("harbor api %s called at %s  cost:%s", ctl, beg.String(), time.Now().Sub(beg).String())
}
