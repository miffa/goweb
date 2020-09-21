package debug

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

func DumpStacks() {
	buf := make([]byte, 1<<16)
	buf = buf[:runtime.Stack(buf, true)]
	log.Info("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", string(buf))
}

func Timestamp() string {
	return time.Now().Format("20060102150405")
}

const (
	WEB_FORMAT = "20060102150405"
)

func Str2Time(str string) (time.Time, error) {
	return time.Parse(WEB_FORMAT, str)
}

const TIME_FORMAT = "2016-01-02 15:04:05"
const TIME_FORMAT_UTC = "2006-01-02T15:04:05.000Z"

func getTime(name string, mp map[string]interface{}) time.Time {

	var retime time.Time
	ctime := GetValueString(name, mp)
	if ctime != "" {
		retime, _ = time.Parse(TIME_FORMAT, ctime)
	}
	return retime
}

var (
	textctrl_u16   = []byte{0xFE, 0xFF}
	textctrl_u8bom = []byte{0xEF, 0xBB, 0xBF}
	textctrl_u32b  = []byte{0x00, 0x00, 0xFE, 0xFF}
	textctrl_u32l  = []byte{0xFE, 0xFF, 0x00, 0x00}
)

func ISEFEE(sql string) string {
	udata2 := []byte(sql[0:2])
	udata3 := []byte(sql[0:3])
	udata4 := []byte(sql[0:4])
	switch {
	case bytes.Equal(textctrl_u8bom, udata3):
		sql = sql[3:]

	case bytes.Equal(textctrl_u16, udata2):
		sql = sql[2:]

	case bytes.Equal(textctrl_u32b, udata4):
		sql = sql[4:]

	case bytes.Equal(textctrl_u32l, udata4):
		sql = sql[4:]
	}
	return sql
}

func FileFuncLine(s bool, dep int) string {
	pc, file, line, ok := runtime.Caller(dep)
	if !ok {
		file = "???"
		line = 0
	}
	if s {
		file = filepath.Base(file)
	}

	funcname := runtime.FuncForPC(pc).Name()
	//funcname = filepath.Ext(funcname)
	return fmt.Sprintf("%s:%s:%d", file, funcname, line)
}
