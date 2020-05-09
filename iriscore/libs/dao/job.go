package dao

import (
	"context"
	"crypto/md5"
	"encoding/hex"
)

var (
	jobmgr map[string]context.CancelFunc
)

func init() {
	jobmgr = make(map[string]context.CancelFunc)
}

func Key(sql string) string {

	dd := md5.Sum([]byte(sql))
	return hex.EncodeToString(dd[:])
}

func AddJob(sql string, c context.CancelFunc) {
	jobmgr[Key(sql)] = c
}

func CancelJob(sql string) bool {
	key := Key(sql)
	jobfunc, ok := jobmgr[key]
	if ok {
		jobfunc()
		delete(jobmgr, key)
		return true
	}
	return false
}
