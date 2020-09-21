package debug

import (
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var D *DebugInfo

func init() {
	D = new(DebugInfo)
	TurnOffDebug()
}

type DebugInfo struct {
	debugok bool
}

func TurnOnDebug() {
	D.debugok = true
}

func TurnOffDebug() {
	D.debugok = false
}

func Log1(format string, v ...interface{}) {
	if D.debugok {
		log.Debugf(format, v...)
	}
}

func Log(format string, v ...interface{}) {

	if D.debugok {
		function, file, line, _ := runtime.Caller(1)
		myformat := "%s:%s:%d:" + format
		var myv []interface{}
		myv = append(myv, chopPath(file))
		myv = append(myv, runtime.FuncForPC(function).Name())
		myv = append(myv, line)
		myv = append(myv, v...)
		log.Debugf(myformat, myv...)
	}
}
func LogDebug(format string, v ...interface{}) {

	function, file, line, _ := runtime.Caller(1)
	myformat := "%s:%s:%d:" + format
	var myv []interface{}
	myv = append(myv, chopPath(file))
	myv = append(myv, runtime.FuncForPC(function).Name())
	myv = append(myv, line)
	myv = append(myv, v...)
	log.Debugf(myformat, myv...)
}

func LogInfo(format string, v ...interface{}) {

	function, file, line, _ := runtime.Caller(1)
	myformat := "%s:%s:%d:" + format
	var myv []interface{}
	myv = append(myv, chopPath(file))
	myv = append(myv, runtime.FuncForPC(function).Name())
	myv = append(myv, line)
	myv = append(myv, v...)
	log.Infof(myformat, myv...)
}

func LogWarn(format string, v ...interface{}) {

	function, file, line, _ := runtime.Caller(1)
	myformat := "%s:%s:%d:" + format
	var myv []interface{}
	myv = append(myv, chopPath(file))
	myv = append(myv, runtime.FuncForPC(function).Name())
	myv = append(myv, line)
	myv = append(myv, v...)
	log.Infof(myformat, myv...)
}

func LogError(format string, v ...interface{}) {
	function, file, line, _ := runtime.Caller(1)
	myformat := "%s:%s:%d:" + format
	var myv []interface{}
	myv = append(myv, chopPath(file))
	myv = append(myv, runtime.FuncForPC(function).Name())
	myv = append(myv, line)
	myv = append(myv, v...)
	log.Errorf(myformat, myv...)
}

func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	} else {
		return original[i+1:]
	}
}

//debug.D.Log()
