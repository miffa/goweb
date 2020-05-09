package perf

import (
	"net/http"
	"net/http/pprof"

	log "goweb/iriscore/libs/logrus"
)

// StartPprof start http pprof.
func Init(pprofBind string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofServeMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	go func() {
		if err := http.ListenAndServe(pprofBind, pprofServeMux); err != nil {
			log.Errorf("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", pprofBind, err)
			panic(err)
		}
	}()
}
