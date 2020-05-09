package dbtimeout

import "time"

var (
	CONNECTION_TIMEOUT = 100 * time.Second
	READ_TIMEOUT       = 10 * time.Second
	WRITE_TIMEOUT      = 10 * time.Second
	IDLE_TIMEOUT       = 10 * time.Second
)

func SetConnTMOUT(d time.Duration) {
	CONNECTION_TIMEOUT = d
}

func SetReadTMOUT(d time.Duration) {
	READ_TIMEOUT = d
}

func SetWriteTMOUT(d time.Duration) {
	WRITE_TIMEOUT = d
}

func SetIdleTMOUT(d time.Duration) {
	IDLE_TIMEOUT = d
}
