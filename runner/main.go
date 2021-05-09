package main

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
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

	ctrl, err := ctrl.StartController(info)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		ctrl.Close()
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

	g := simple.NewWeightedUndirectedGraph(0, 0)
	a, b, c := simple.Node(0), simple.Node(1), simple.Node(2)
	g.AddNode(a)
	g.AddNode(b)
	g.AddNode(c)

	g.SetWeightedEdge(g.NewWeightedEdge(a, b, 1))
	g.SetWeightedEdge(g.NewWeightedEdge(a, c, 1))
	g.SetWeightedEdge(g.NewWeightedEdge(b, c, 1))

	fmt.Println("starting processes")
	err = ctrl.StartProcesses(cfg, g, &brb.Flooding{}, 1, []uint64{0})
	if err != nil {
		fmt.Printf("unable to start processes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("waiting for all process to be alive")
	if err := ctrl.WaitForAlive(); err != nil {
		fmt.Printf("err while waiting for alive: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("waiting for all process to be ready")
	if err := ctrl.WaitForReady(); err != nil {
		fmt.Printf("err while waiting for ready: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("everything ready, sending test msg")
	time.Sleep(time.Second)
	uid, err := ctrl.TriggerMessageSend(0, []byte("blah"))
	if err != nil {
		fmt.Printf("err while sending payload msg: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("sent message (%v), waiting for deliver\n", uid)
	stats := ctrl.WaitForDeliver(uid)
	fmt.Printf("stats: %v\n", stats)

	<-stopCh
	fmt.Println("server stop")
}
