package main

import (
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"os"
	"os/signal"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
	"syscall"
	"time"
)

func init() {
	gob.Register(simple.WeightedEdge{})
	gob.Register(simple.Node(0))
}

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
		Sock: "ipc:///tmp/rp-ctl.ipc",
	}

	ctl, err := ctrl.StartController(info)
	if err != nil {
		fmt.Printf("unable to start controller: %v\n", err)
		os.Exit(1)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		ctl.Close()
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

	n, k, f := 10, 3, 1
	m := graphs.GeneralizedWheelGenerator{}
	if err := runSimpleTest(ctl, 15, n, k, f, m, cfg, &brb.DolevImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done")
	<-stopCh
	fmt.Println("server stop")
}

func runSimpleTest(ctl *ctrl.Controller, runs int, n, k, f int, gen graphs.Generator, cfg process.Config, bp brb.Protocol) error {
	if k < 2*f+1 {
		return errors.Errorf("network is not 2f+1 connected (k=%v, f=%v)", k, f)
	}

	g, err := gen.Generate(n, k)
	if err != nil {
		return errors.Wrap(err, "failed to generate graph for test")
	}

	fmt.Println("starting processes")
	err = ctl.StartProcesses(cfg, g, bp, f, []uint64{0})
	if err != nil {
		return errors.Wrap(err, "unable to start processes")
	}

	fmt.Println("waiting for all process to be alive")
	if err := ctl.WaitForAlive(); err != nil {
		return errors.Wrap(err, "err while waiting for alive")
	}

	fmt.Println("waiting for all process to be ready")
	if err := ctl.WaitForReady(); err != nil {
		return errors.Wrap(err, "err while waiting for ready")
	}

	lat := time.Duration(0)
	lats := make([]time.Duration, 0, runs)
	msg := 0

	fmt.Printf("everything ready, starting %v test runs\n", runs)
	for i := 0; i < runs; i++ {
		uid, err := ctl.TriggerMessageSend(0, []byte("blah"))
		if err != nil {
			fmt.Printf("err while sending payload msg: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("sent message (%v, round %v), waiting for deliver\n", uid, i)
		stats := ctl.WaitForDeliver(uid)
		fmt.Printf("statistics (%v):\n  last delivery latency: %v\n  messages sent: %v\n", i,
			stats.Latency, stats.MsgCount)
		lat += stats.Latency
		lats = append(lats, stats.Latency)
		msg += stats.MsgCount
	}

	fmt.Println("average stats:")
	mean, sd := sd(lats)
	rsd := float64(sd) * 100 / float64(mean)
	fmt.Printf("  latency:\n    mean: %v\n    sd: %v (%.2f%%)\n", mean, sd, rsd)
	fmt.Printf("  messages: %v\n", msg/runs)

	fmt.Println("config:")
	fmt.Printf("  nodes: %v\n  connectivity (k): %v\n  byzantine nodes (f): %v\n  runs: %v\n", n, k, f, runs)

	return nil
}

func sd(xs []time.Duration) (time.Duration, time.Duration) {
	if len(xs) < 2 {
		return xs[0], 0
	}

	sum := time.Duration(0)
	for _, i := range xs {
		sum += i
	}

	m := sum / time.Duration(len(xs))

	diff := time.Duration(0)
	for _, i := range xs {
		diff += (i - m) * (i - m)
	}

	v := diff / time.Duration(len(xs)-1)

	return m, time.Duration(math.Sqrt(float64(v)))
}
