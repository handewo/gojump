package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/handewo/gojump/pkg/server"
	"github.com/sevlyar/go-daemon"
)

func runDaemon() {
	ctx := &daemon.Context{
		PidFileName: pidPath,
		PidFilePerm: 0644,
		Umask:       027,
	}
	child, err := ctx.Reborn()
	if err != nil {
		log.Fatalf("run failed: %v", err)
	}
	if child != nil {
		return
	}
	defer ctx.Release()
	server.Run(configPath)
}

var (
	pidPath = "gojump.pid"

	daemonFlag  = false
	stopFlag    = false
	versionFlag = false

	configPath = ""
)

func init() {
	flag.BoolVar(&daemonFlag, "d", false, "run as Daemon")
	flag.BoolVar(&stopFlag, "s", false, "stop service")
	flag.StringVar(&configPath, "f", "config.yml", "config path")
	flag.StringVar(&pidPath, "p", "gojump.pid", "pid path")
	flag.BoolVar(&versionFlag, "v", false, "version")
}

func main() {
	flag.Parse()
	if versionFlag {
		fmt.Printf("Version:             %s\n", server.Version)
		return
	}

	if stopFlag {
		pid, err := ioutil.ReadFile(pidPath)
		if err != nil {
			log.Fatal("Pid file not exist")
			return
		}
		pidInt, _ := strconv.Atoi(string(pid))
		err = syscall.Kill(pidInt, syscall.SIGTERM)
		if err != nil {
			log.Fatalf("Stop failed: %v", err)
		} else {
			_ = os.Remove(pidPath)
		}
		return
	}

	if daemonFlag {
		runDaemon()
	} else {
		server.Run(configPath)
	}
}
