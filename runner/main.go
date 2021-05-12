package main

import (
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
	"syscall"
	"time"

	_ "net/http/pprof"
)

func init() {
	gob.Register(simple.WeightedEdge{})
	gob.Register(simple.Node(0))
	gob.Register(graphs.Node{})
}

func main() {
	//graphs.GraphsMain()
	RunnerMain()
}

// TODO: keep in mind high water mark
func RunnerMain() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("starting rp runner")
	stopCh := make(chan struct{}, 1)
	info := ctrl.ControllerInfo{
		ID:   "RP-CONTROLLER",
		Sock: "ipc:///tmp/rp-ctl.ipc",
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
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

	n, k, f := 100, 5, 0
	m := graphs.GeneralizedWheelGenerator{}
	if err := runSimpleTest(info, 5, n, k, f, m, cfg, &brb.DolevKnown{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done")
	<-stopCh
	fmt.Println("server stop")
}

func runSimpleTest(info ctrl.ControllerInfo, runs int, n, k, f int, gen graphs.Generator, cfg process.Config, bp brb.Protocol) error {
	if k < 2*f+1 {
		return errors.Errorf("network is not 2f+1 connected (k=%v, f=%v)", k, f)
	}

	g, err := gen.Generate(n, k)
	if err != nil {
		return errors.Wrap(err, "failed to generate graph for test")
	}

	fmt.Printf("everything ready, starting %v test runs\n", runs)

	ctl, err := ctrl.StartController(info)
	if err != nil {
		return errors.Wrap(err, "unable to start controller")
	}

	fmt.Println("starting processes")
	err = ctl.StartProcesses(cfg, g, bp, f, []uint64{0, 1})
	if err != nil {
		return errors.Wrap(err, "unable to start processes")
	}

	lat := time.Duration(0)
	lats := make([]int, 0, runs)
	msg := 0
	cnts := make([]int, 0, runs)

	for i := 0; i < runs; i++ {
		fmt.Printf("---\nrun %v: waiting for all process to be alive\n", i)
		if err := ctl.WaitForAlive(); err != nil {
			return errors.Wrap(err, "err while waiting for alive")
		}

		fmt.Printf("run %v: waiting for all process to be ready\n", i)
		if err := ctl.WaitForReady(); err != nil {
			return errors.Wrap(err, "err while waiting for ready")
		}

		uid1, err := ctl.TriggerMessageSend(0, []byte(fmt.Sprintf("run_%v", i)))
		if err != nil {
			fmt.Printf("err while sending payload msg: %v\n", err)
			os.Exit(1)
		}

		uid2, err := ctl.TriggerMessageSend(0, []byte(fmt.Sprintf("run_%v", i)))
		if err != nil {
			fmt.Printf("err while sending payload msg: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("sent message (%v & %v, round %v), waiting for deliver\n", uid1, uid2, i)
		stats := ctl.WaitForDeliver(uid1)
		fmt.Printf("statistics (%v, %v):\n  last delivery latency: %v\n  messages sent: %v\n", uid1, i,
			stats.Latency, stats.MsgCount)
		stats2 := ctl.WaitForDeliver(uid2)
		fmt.Printf("statistics (%v, %v):\n  last delivery latency: %v\n  messages sent: %v\n", uid2, i,
			stats2.Latency, stats2.MsgCount)

		lat += stats.Latency
		lats = append(lats, int(stats.Latency))
		msg += stats.MsgCount
		cnts = append(cnts, stats.MsgCount)

		time.Sleep(time.Second * 2)
	}

	fmt.Println("average stats:")
	lMean, lSd := sd(lats)
	lRsd := lSd * 100 / lMean
	fmt.Printf("  latency:\n    mean: %v\n    sd: %v (%.2f%%)\n", time.Duration(lMean), time.Duration(lSd), lRsd)

	mMean, mSd := sd(cnts)
	mRsd := mSd * 100 / mMean
	fmt.Printf("  messages:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", mMean, mSd, mRsd)

	fmt.Println("config:")
	fmt.Printf("  nodes: %v\n  connectivity (k): %v\n  byzantine nodes (f): %v\n  runs: %v\n", n, k, f, runs)

	return nil
}

func sd(xs []int) (float64, float64) {
	if len(xs) < 2 {
		return float64(xs[0]), 0
	}

	sum := 0
	for _, i := range xs {
		sum += i
	}

	m := float64(sum) / float64(len(xs))

	diff := float64(0)
	for _, i := range xs {
		diff += (float64(i) - m) * (float64(i) - m)
	}

	v := diff / float64(len(xs)-1)

	return m, math.Sqrt(v)
}
