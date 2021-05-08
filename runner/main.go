package main

import (
	"fmt"
	"os"
	"os/signal"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/process"
	"syscall"
	"time"
)

func main() {
	//graphs.GraphsMain()
	RunnerMain()
}

// TODO: keep in mind high water mark
func RunnerMain() {
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
		CtrlID:         info.ID,
		CtrlSock:       info.Sock,
		Sock:           "ipc:///tmp/rp-node-%v.ipc",
		MaxRetries:     5,
		RetryDelay:     time.Millisecond * 100,
		NeighbourDelay: time.Millisecond * 300,
	}

	fmt.Println("starting process 42")
	err = c.StartProcess(42, cfg, []uint16{4242}, &brb.Flooding{}, false)
	if err != nil {
		fmt.Printf("unable to start process: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("starting process 4242")
	err = c.StartProcess(4242, cfg, []uint16{42}, &brb.Flooding{}, false)
	if err != nil {
		fmt.Printf("unable to start process: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("waiting for all process to be alive")
	if err := c.WaitForAlive(); err != nil {
		fmt.Printf("err while waiting for alive: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("waiting for all process to be ready")
	if err := c.WaitForReady(); err != nil {
		fmt.Printf("err while waiting for ready: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("everything ready, sending test msg")
	time.Sleep(time.Second)
	uid, err := c.TriggerMessageSend(42, []byte("blah"))
	if err != nil {
		fmt.Printf("err while sending payload msg: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("sent message (%v), waiting for deliver\n", uid)
	stats := c.WaitForDeliver(uid)
	fmt.Printf("stats: %v\n", stats)

	<-stopCh
	fmt.Println("server stop")
}
