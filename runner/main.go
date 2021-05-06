package main

import (
	"fmt"
	"os"
	"os/signal"
	"rp-runner/ctrl"
	"rp-runner/process"
	"syscall"
	"time"
)

func main() {
	fmt.Println("starting rp runner")
	stopCh := make(chan struct{}, 1)
	info := ctrl.ControllerInfo{
		ID:   "RP-CONTROLLER",
		Sock: "ipc:///tmp/rp-ctrl.ipc",
	}

	c, err := ctrl.StartController(info)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		c.Close()
		close(stopCh)
	}()

	cfg := process.Config{
		CtrlID:     info.ID,
		CtrlSock:   info.Sock,
		Sock:       "ipc:///tmp/rp-node-%v.ipc",
		MaxRetries: 5,
		RetryDelay: time.Millisecond * 100,
	}

	fmt.Println("starting process 42")
	_, err = process.StartProcess(42, cfg, stopCh)
	if err != nil {
		fmt.Printf("unable to start process: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("starting process 4242")
	_, err = process.StartProcess(4242, cfg, stopCh)
	if err != nil {
		fmt.Printf("unable to start process: %v\n", err)
		os.Exit(1)
	}

	<-stopCh
	fmt.Println("server stop")
}
