package main

import (
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
	"runtime"
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

func RunnerMain() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("starting rp runner")
	stopCh := make(chan struct{}, 1)
	info := ctrl.Config{
		PollDelay:  time.Millisecond * 200,
		CtrlBuffer: 2000,
		ProcBuffer: 50000,
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
		MaxRetries:     5,
		RetryDelay:     time.Millisecond * 100,
		NeighbourDelay: time.Millisecond * 300,
	}

	// TEMP
	//gr := simple.NewWeightedUndirectedGraph(0, 0)
	////Create nodes
	//a := graphs.NewNodeUndirected(gr, "a")
	//gr.AddNode(a)
	//b := graphs.NewNodeUndirected(gr, "b")
	//gr.AddNode(b)
	//c := graphs.NewNodeUndirected(gr, "c")
	//gr.AddNode(c)
	//d := graphs.NewNodeUndirected(gr, "d")
	//gr.AddNode(d)
	//e := graphs.NewNodeUndirected(gr, "e")
	//gr.AddNode(e)
	//f := graphs.NewNodeUndirected(gr, "f")
	//gr.AddNode(f)
	//g := graphs.NewNodeUndirected(gr, "g")
	//gr.AddNode(g)
	//
	//ab := gr.NewWeightedEdge(a, b, 1)
	//ac := gr.NewWeightedEdge(a, c, 1)
	//ad := gr.NewWeightedEdge(a, d, 1)
	//bc := gr.NewWeightedEdge(b, c, 1)
	//dc := gr.NewWeightedEdge(d, c, 1)
	//be := gr.NewWeightedEdge(b, e, 1)
	//ce := gr.NewWeightedEdge(c, e, 1)
	//cf := gr.NewWeightedEdge(c, f, 1)
	//cg := gr.NewWeightedEdge(c, g, 1)
	//ef := gr.NewWeightedEdge(e, f, 1)
	//fg := gr.NewWeightedEdge(f, g, 1)
	//dg := gr.NewWeightedEdge(d, g, 1)
	//
	//gr.SetWeightedEdge(ab)
	//gr.SetWeightedEdge(ac)
	//gr.SetWeightedEdge(ad)
	//gr.SetWeightedEdge(bc)
	//gr.SetWeightedEdge(dc)
	//gr.SetWeightedEdge(be)
	//gr.SetWeightedEdge(ce)
	//gr.SetWeightedEdge(cf)
	//gr.SetWeightedEdge(cg)
	//gr.SetWeightedEdge(ef)
	//gr.SetWeightedEdge(fg)
	//gr.SetWeightedEdge(dg)

	// Optimizations
	opts := brb.OptimizationConfig{
		DolevFilterSubpaths:         false,
		DolevSingleHopNeighbour:     false, // orbd.2
		DolevCombineNextHops:        false, // orbd.2
		DolevReusePaths:             false,
		DolevRelayMerging:           false,
		DolevPayloadMerging:         false,
		DolevImplicitPath:           false,
		BrachaImplicitEcho:          false,
		BrachaMinimalSubset:         false,
		BrachaDolevPartialBroadcast: false,
		BrachaDolevMerge:            false,
	}

	//n, k, fx := 25, 8, 2
	//messages := 5
	//deg := k
	payloadSize := 12000
	//gen := graphs.RandomRegularGenerator{}
	//_, name := gen.Cache()
	//
	//cache := graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	//if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
	//	fmt.Printf("err while running simple test: %v\n", err)
	//	os.Exit(1)
	//}

	brachaIndividualTests(opts, info, cfg, payloadSize, false)

	fmt.Println("done")
	fmt.Println("server stop")
}

func generatePayload(size, run int) bytePayload {
	if size <= 1 {
		return bytePayload(fmt.Sprintf("%v", run))
	}

	payload := ""
	for i := 2; i < size; i++ {
		payload += "X"
	}

	return bytePayload(fmt.Sprintf("%v_%v", payload, run))
}

func pickRandom(i int, max int) []uint64 {
	res := make([]uint64, 0, i)
	used := make(map[uint64]struct{})
	overlap := i > max

	for len(res) < i {
		rand.Seed(time.Now().UnixNano())
		next := uint64(rand.Intn(max))

		if _, ok := used[next]; !overlap && ok {
			continue
		}

		used[next] = struct{}{}
		res = append(res, next)
	}

	return res
}

type bytePayload []byte

func (b bytePayload) SizeOf() uintptr {
	return uintptr(len(b))
}

func runMultipleMessagesTest(info ctrl.Config, runs int, n, k, f, deg, messages, payloadSize int, gen graphs.Generator, cfg process.Config, opt brb.OptimizationConfig, bp brb.Protocol) error {
	if k < 2*f+1 && bp.Category() != brb.BrachaCat {
		return errors.Errorf("network is not 2f+1 connected (k=%v, f=%v)", k, f)
	}

	if float64(f) >= float64(n)/3 && bp.Category() != brb.DolevCat {
		return errors.Errorf("f >= n/3 (n=%v, f=%v)", n, f)
	}

	ra := pickRandom(runs*messages, n-f)
	g, err := gen.Generate(n, k, deg)
	if err != nil {
		return errors.Wrap(err, "failed to generate graph for test")
	}

	fmt.Printf("everything ready, starting %v test runs\n", runs)

	ctl, err := ctrl.StartController(info)
	if err != nil {
		return errors.Wrap(err, "unable to start controller")
	}

	fmt.Printf("starting processes\nselected as possible transmitters: %v\n", ra)
	err = ctl.StartProcesses(cfg, opt, g, bp, f, ra, bp.Category() == brb.BrachaDolevCat)
	if err != nil {
		return errors.Wrap(err, "unable to start processes")
	}

	lats := make([]int, 0, runs)
	cnts := make([]int, 0, runs)
	bdMergeds := make([]int, 0, runs)
	dMergeds := make([]int, 0, runs)
	pMergeds := make([]int, 0, runs)
	transmits := make([]int, 0, runs)

	for i := 0; i < runs; i++ {
		fmt.Printf("---\nrun %v: waiting for all process to be alive\n", i)
		if err := ctl.WaitForAlive(); err != nil {
			return errors.Wrap(err, "err while waiting for alive")
		}

		fmt.Printf("run %v: waiting for all process to be ready\n", i)
		if err := ctl.WaitForReady(); err != nil {
			return errors.Wrap(err, "err while waiting for ready")
		}

		uids := make([]uint32, 0, messages)
		payload := generatePayload(payloadSize, i)
		for j := 0; j < messages; j++ {
			id := ra[i*messages+j]

			uid, err := ctl.TriggerMessageSend(id, payload)
			if err != nil {
				fmt.Printf("err while sending payload msg: %v\n", err)
				os.Exit(1)
			}

			uids = append(uids, uid)
		}

		fmt.Printf("sent %v messages (%v, round %v, origins %v) of %v bytes, waiting for delivers\n", messages, uids,
			i, ra[i*messages:i*messages+messages], payload.SizeOf())

		roundLat := time.Duration(0)
		roundMsg := 0
		roundBDMerged := 0
		roundDMerged := 0
		roundPMerged := 0
		roundRelayCnt := 0
		roundMinRelayCnt := math.MaxInt64
		roundMaxRelayCnt := 0
		roundMeanRelayCnt := 0.0
		roundTransmitted := 0
		for _, uid := range uids {
			stats := ctl.WaitForDeliver(uid)

			if stats.Latency > roundLat {
				roundLat = stats.Latency
			}

			if stats.MinRelayCnt < roundMinRelayCnt {
				roundMinRelayCnt = stats.MinRelayCnt
			}

			if stats.MaxRelayCnt > roundMaxRelayCnt {
				roundMaxRelayCnt = stats.MaxRelayCnt
			}

			roundMsg += stats.MsgCount
			roundRelayCnt += stats.RelayCnt
			roundMeanRelayCnt += stats.MeanRelayCount
			roundBDMerged += stats.BDMessagedMerged
			roundTransmitted += stats.BytesTransmitted
			roundDMerged += stats.DMessagesMerged
			roundPMerged += stats.PayloadsMerged
		}

		roundMeanRelayCnt /= float64(messages)

		fmt.Printf("statistics (%v):\n  last delivery latency: %v\n  messages sent: %v (~%v per broadcast)"+
			"\n  recv: %.2f (%v - %v - %v)\n  bd merged: %v\n  d merged: %v\n  payloads merged: %v\n  "+
			"bytes transmitted: %v (~%v per broadcast)\n", i,
			roundLat, roundMsg, roundMsg/messages, roundMeanRelayCnt, roundRelayCnt, roundMinRelayCnt, roundMaxRelayCnt,
			roundBDMerged, roundDMerged, roundPMerged, roundTransmitted, roundTransmitted/messages)

		lats = append(lats, int(roundLat))
		cnts = append(cnts, roundMsg/messages)
		bdMergeds = append(bdMergeds, roundBDMerged)
		dMergeds = append(dMergeds, roundDMerged)
		pMergeds = append(pMergeds, roundPMerged)
		transmits = append(transmits, roundTransmitted/messages)

		ctl.FlushProcesses()
		runtime.GC()
	}

	fmt.Println("==========\naverage stats:")
	lMean, lSd := sd(lats)
	lRsd := lSd * 100 / lMean
	fmt.Printf("  latency:\n    mean: %v\n    sd: %v (%.2f%%)\n", time.Duration(lMean), time.Duration(lSd), lRsd)

	mMean, mSd := sd(cnts)
	mRsd := mSd * 100 / mMean
	fmt.Printf("  messages:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", mMean, mSd, mRsd)

	bdmMean, bdmSd := sd(bdMergeds)
	bdmRsd := bdmSd * 100 / bdmMean
	fmt.Printf("  bd merged:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", bdmMean, bdmSd, bdmRsd)

	dmMean, dmSd := sd(dMergeds)
	dmRsd := dmSd * 100 / dmMean
	fmt.Printf("  d merged:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", dmMean, dmSd, dmRsd)

	pmMean, pmSd := sd(pMergeds)
	pmRsd := pmSd * 100 / pmMean
	fmt.Printf("  payloads merged:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", pmMean, pmSd, pmRsd)

	tMean, tSd := sd(transmits)
	tRsd := tSd * 100 / tMean
	fmt.Printf("  transmits:\n    mean: %.2f (~%.2f per broadcast)\n    sd: %.2f (%.2f%%)\n", tMean,
		tMean/float64(messages), tSd, tRsd)

	fmt.Println("config:")
	fmt.Printf("  nodes: %v\n  connectivity (k): %v\n  byzantine nodes (f): %v"+
		"\n  runs: %v\n  protocol: %v\n  payload size: %v bytes\n  messages: %v\n",
		n, k, f, runs, reflect.TypeOf(bp).Elem().Name(), payloadSize, messages)

	ctl.FlushProcesses()
	ctl.Close()

	fmt.Println("==========")
	runtime.GC()

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
