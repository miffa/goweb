package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"goweb/iriscore/config"
	"goweb/iriscore/iocgo"
	"goweb/iriscore/router"
	"goweb/iriscore/version"

	"goweb/iriscore/libs/debug"
	"goweb/iriscore/libs/perf"

	log "goweb/iriscore/libs/logrus"

	// register ioc
	_ "goweb/iriscore/middleware/tracinglog"
	_ "goweb/iriscore/resource"
	_ "goweb/iriscore/service"
	_ "goweb/iriscore/thirdurl"
)

const (
	serviceName = `tpaasportal`
	HttpTimeout = 30
)

var cfgfile = flag.String("c", "config.yaml", "configuration file, default to config.yaml")
var ver = flag.Bool("version", false, "Output version and exit")

func main() {
	// args parse
	flag.Parse()

	version.Service = serviceName

	if *ver {
		fmt.Println(version.Service, ": ", version.Version)
		return
	}

	// config file parse
	cfg, err := config.NewTpaasConfig(*cfgfile)
	if err != nil {
		fmt.Printf(" load conf %s err:%v", *cfgfile, err)
		return
	}

	//cfggstr, _ := json.Marshal(cfg.GetAllConfig())
	//fmt.Printf("---%s\n", string(cfggstr))

	// logger
	rf := log.NewRotateFile(cfg.GetString("common.service_log"), 100*log.MiB)
	defer rf.Close()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(rf)

	//  log level
	log.SetLevel(log.DebugLevel)
	switch cfg.GetString("common.log_level") {
	case "ERROR":
		fallthrough
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		fallthrough
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "INFO":
		fallthrough
	case "info":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		fallthrough
	case "debug":
		log.SetLevel(log.DebugLevel)
	}

	//pid file
	err = InitPidfile(cfg)
	if err != nil {
		fmt.Printf("initpidfile err:%v", err)
		return

	}
	defer QuitPidFile(cfg)

	//////////////////////////////////////
	//     service init
	//  init the dependency service
	err = InitDependencyService(cfg)
	if err != nil {
		log.Errorf("init dependency service err:%v", err)
		fmt.Printf("init dependency service err:%v", err)
		return
	}
	log.Infof("init dependency service ok")

	//  when service stopping, close the dependency service
	defer CloseDependencyService()

	///////////////////////////////////////
	////  http service
	// perf service
	if cfg.IsSet("http.pprof_addr") {
		perf.Init(cfg.GetString("http.pprof_addr"))
		log.Infof("http pprof service init ok")
	}

	if !cfg.IsSet("http.http_addr") {
		log.Errorf("http address is not in config")
		fmt.Printf("http address is not in config")
		return
	}

	httptimeout := cfg.GetInt("http.http_timeout")
	if httptimeout == 0 {
		httptimeout = HttpTimeout
	}

	// http service init
	router.Api(). // singleTon api
			ConfigDefault().
			SetTimeout(time.Duration(httptimeout) * time.Second).
			SetLog(rf).
			InitRouter().
			Runapi(cfg.GetString("http.http_addr"))
	log.Infof("http service init ok")

	////////
	log.Infof("          ________                                                     ")
	log.Infof("       __/_/      |______   %s.%s is running                ", version.Service, version.Version)
	log.Infof("      / O O O O O O O O O ...........................................  ")
	log.Infof("                                                                       ")
	log.Infof("      %s", time.Now().String())
	log.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	////////
	// signal
	InitSignal()
}

func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP, syscall.SIGUSR1, syscall.SIGUSR2)
	//log.Infof("ait for signal.......")
	for {
		s := <-c
		log.Infof("service[%s] get a signal %s", version.Version, s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT, syscall.SIGHUP:
			GracefulQuit()
			return
		case syscall.SIGUSR2:
			debug.DumpStacks()
		case syscall.SIGUSR1:
			// todo: your operation
			//return
		default:
			//return
		}
	}
}

func GracefulQuit() {
	log.Infof("service make a graceful quit !!!!!!!!!!!!!!")
	router.Api().Shutdown() // close http service
	// close your service here

	time.Sleep(1 * time.Second)
}

func InitDependencyService(cfg *config.TpaasConfig) error {
	return iocgo.LaunchEngine(cfg)
}

func CloseDependencyService() error {
	return iocgo.StopEngine()
}

func InitPidfile(cfg *config.TpaasConfig) error {
	//pid file
	pidfile := ""
	if !cfg.IsSet("common.pid_file") {
		return nil
	} else {
		pidfile = cfg.GetString("common.pid_file")
	}
	contents, err := ioutil.ReadFile(pidfile)
	if err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
		if err != nil {
			log.Errorf("Error reading proccess id from pidfile '%s': %s",
				pidfile, err)
			return errors.WithMessage(err, "reading proccess id from pidfile")
		}

		process, err := os.FindProcess(pid)

		// on Windows, err != nil if the process cannot be found
		if runtime.GOOS == "windows" {
			if err == nil {
				log.Errorf("Process %d is already running.", pid)
				return fmt.Errorf("already running")
			}
		} else if process != nil {
			// err is always nil on POSIX, so we have to send the process
			// a signal to check whether it exists
			if err = process.Signal(syscall.Signal(0)); err == nil {
				log.Errorf("Process %d is already running.", pid)
				return fmt.Errorf("already running")
			}
		}
	}
	if err = ioutil.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())),
		0644); err != nil {

		log.Errorf("Unable to write pidfile '%s': %s", pidfile, err)
		return err
	}
	log.Infof("Wrote pid to pidfile '%s'", pidfile)
	return nil
}

func QuitPidFile(cfg *config.TpaasConfig) {
	pidfile := ""
	if !cfg.IsSet("common.pid_file") {
		return
	} else {
		pidfile = cfg.GetString("common.pid_file")
	}

	if err := os.Remove(pidfile); err != nil {
		log.Errorf("Unable to remove pidfile '%s': %s", pidfile, err)
	}
	return
}
