package main

import (
	"encoding/gob"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
	"runtime"
	"time"

	_ "net/http/pprof"
)

func init() {
	gob.Register(simple.WeightedEdge{})
	gob.Register(simple.Node(0))
	gob.Register(graphs.Node{})
}

// This function is the main program entry
func main() {
	val, ok := os.LookupEnv("MANUAL_RUNNER")
	if ok && val == "true" {
		fmt.Println("running manual runner")
		RunnerMain()
	} else {
		CLI()
	}
}

// You can test in this function
func RunnerMain() {
	// Profiler (pprof) is started in the background, in case you're interested in detailed performance reports
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("starting rp runner")
	// You can change these if you want, but these values should be sane defaults
	info := ctrl.Config{
		// Controller uses polling for waiting etc
		PollDelay: time.Millisecond * 200,

		// Buffer size of controller
		CtrlBuffer: 2000,

		// Buffer size of each link. Be aware that the full buffer is allocated, even though it might not be used.
		ProcBuffer: 50000,

		// Set the verbosity
		Verbosity: ctrl.SLOW,
	}
	cfg := process.Config{
		// Connection attempts to controller and neighbours
		MaxRetries: 5,

		// Timeout delays
		RetryDelay:     time.Millisecond * 100,
		NeighbourDelay: time.Millisecond * 300,
	}

	// Optimizations
	opts := brb.OptimizationConfig{
		// All optimizations are ordered in their ORX.Y order
		DolevFilterSubpaths:         true,
		DolevSingleHopNeighbour:     true,
		DolevCombineNextHops:        true,
		DolevReusePaths:             true,
		DolevRelayMerging:           true,
		DolevPayloadMerging:         true,
		DolevImplicitPath:           true,
		BrachaImplicitEcho:          true,
		BrachaMinimalSubset:         true,
		BrachaDolevPartialBroadcast: true,
		BrachaDolevMerge:            true,
	}

	// Run a single test
	runCfg := RunConfig{
		// Amount of times to run the test (used for standard deviation + mean)
		Runs: 5,

		// Amount of nodes
		N: 25,

		// Connectivity
		K: 8,

		// Amount of Byzantine nodes
		F: 3,

		// Size of the generated payload
		PayloadSize: 12,

		// Boolean indicating if all non-Byzantine nodes will transmit (the same payload)
		MultipleTransmitters: false,

		// Graph generator to use. Possible options
		// - RandomRegularGenerator
		// - GeneralizedWheelGenerator (not complete, will place Byzantine nodes randomly, not yet in the center)
		// - MultiPartiteWheelGenerator / MultiPartiteWheelAltGenerator (same graph, different implementation)
		// - FullyConnectedGenerator
		// - BasicGenerator (will return given graph, used for testing specific graphs)
		Generator: graphs.RandomRegularGenerator{},

		// Protocols can be:
		// - DolevKnownImproved
		// - BrachaImproved
		// - BrachaDolevKnownImproved
		// Others have been used for testing, but are not updated so might not work anymore
		Protocol: &brb.DolevKnownImproved{},

		// The rest can be ignored
		ControlCfg:      info,
		ProcessCfg:      cfg,
		OptimizationCfg: opts,
	}

	// If you want to use the graph cache, enable it.
	// This is recommended for large random graphs, as it can take a few minutes to generate them
	useCache := false
	if useCache {
		_, name := runCfg.Generator.Cache()
		runCfg.Generator = &graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: runCfg.Generator}
	}

	if err := runMultipleMessagesTest(runCfg, false); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	// Run a full test
	brachaFullTests(opts, info, cfg, 12, 5, true, 0)

	fmt.Println("done")
	fmt.Println("server stop")
}

// ==== NOT NEEDED TO LOOK FURTHER FOR SIMPLE TESTING ====
func generatePayload(size, run int) bytePayload {
	if size <= 0 {
		return bytePayload{}
	}

	if size == 1 {
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

type RunConfig struct {
	Runs, N, K, F, Degree, PayloadSize int
	MultipleTransmitters               bool
	Generator                          graphs.Generator
	ControlCfg                         ctrl.Config
	ProcessCfg                         process.Config
	OptimizationCfg                    brb.OptimizationConfig
	Protocol                           brb.Protocol
}

func runMultipleMessagesTest(runCfg RunConfig, skip bool) error {
	if skip {
		return nil
	}

	if runCfg.PayloadSize < 0 {
		return errors.Errorf("invalid payload size, must be >=0 but is: %v", runCfg.PayloadSize)
	}

	if runCfg.N < 0 {
		return errors.Errorf("invalid amount of nodes, must be >=0 but is: %v", runCfg.N)
	}

	if runCfg.F < 0 {
		return errors.Errorf("invalid amount of byzantine nodes, must be >=0 but is: %v", runCfg.F)
	}

	if runCfg.Degree <= 0 {
		runCfg.Degree = runCfg.K
	}

	testGen := runCfg.Generator
	if v, ok := testGen.(*graphs.FileCacheGenerator); ok {
		testGen = v.Gen
	}

	if _, ok := testGen.(*graphs.FullyConnectedGenerator); runCfg.Protocol.Category() == brb.BrachaCat && !ok {
		return errors.New("pure bracha cannot function on non-fully connected networks!")
	}

	if runCfg.Protocol.Category() == brb.BrachaCat {
		runCfg.K = runCfg.N
	}

	if runCfg.K < 2*runCfg.F+1 && runCfg.Protocol.Category() != brb.BrachaCat {
		return errors.Errorf("network is not 2f+1 connected (k=%v, f=%v)", runCfg.K, runCfg.F)
	}

	if float64(runCfg.F) >= float64(runCfg.N)/3 && runCfg.Protocol.Category() != brb.DolevCat {
		return errors.Errorf("f >= n/3 (n=%v, f=%v)", runCfg.N, runCfg.F)
	}

	messages := 1
	if runCfg.MultipleTransmitters {
		messages = runCfg.N - runCfg.F
	}

	fmt.Println("generating graph...")
	ra := pickRandom(runCfg.Runs*messages, runCfg.N-runCfg.F)
	g, err := runCfg.Generator.Generate(runCfg.N, runCfg.K, runCfg.Degree)
	if err != nil {
		return errors.Wrap(err, "failed to generate graph for test")
	}

	fmt.Printf("everything ready, starting %v test runs\n", runCfg.Runs)

	ctl, err := ctrl.StartController(runCfg.ControlCfg)
	if err != nil {
		return errors.Wrap(err, "unable to start controller")
	}

	if runCfg.ControlCfg.Verbosity > ctrl.SILENT {
		fmt.Printf("starting processes\nselected as possible transmitters: %v\n", ra)
	}
	err = ctl.StartProcesses(runCfg.ProcessCfg, runCfg.OptimizationCfg, g, runCfg.Protocol, runCfg.F, ra, runCfg.Protocol.Category() == brb.BrachaDolevCat)
	if err != nil {
		return errors.Wrap(err, "unable to start processes")
	}

	lats := make([]int, 0, runCfg.Runs)
	cnts := make([]int, 0, runCfg.Runs)
	bdMergeds := make([]int, 0, runCfg.Runs)
	dMergeds := make([]int, 0, runCfg.Runs)
	pMergeds := make([]int, 0, runCfg.Runs)
	transmits := make([]int, 0, runCfg.Runs)

	for i := 0; i < runCfg.Runs; i++ {
		fmt.Printf("---\nrun %v: waiting for all process to be alive\n", i)
		if err := ctl.WaitForAlive(); err != nil {
			return errors.Wrap(err, "err while waiting for alive")
		}

		fmt.Printf("run %v: waiting for all process to be ready\n", i)
		if err := ctl.WaitForReady(); err != nil {
			return errors.Wrap(err, "err while waiting for ready")
		}

		uids := make([]uint32, 0, messages)
		payload := generatePayload(runCfg.PayloadSize, i)
		for j := 0; j < messages; j++ {
			id := ra[i*messages+j]

			uid, err := ctl.TriggerMessageSend(id, payload)
			if err != nil {
				fmt.Printf("err while sending payload msg: %v\n", err)
				os.Exit(1)
			}

			uids = append(uids, uid)
		}

		color.Black("sent %v messages (%v, round %v, origins %v) of %v bytes, waiting for delivers\n", messages, uids,
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

		color.Green("statistics (%v):\n  last delivery latency: %v\n  messages sent: %v (~%v per broadcast)"+
			"\n  recv: %.2f (%v - %v - %v)\n  bd merged (orbd2): %v\n  d merged (ord5): %v\n  payloads merged (ord6): %v\n  "+
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

	color.Black("==========\n")
	color.Green("average stats:\n")
	lMean, lSd := sd(lats)
	lRsd := lSd * 100 / lMean
	color.Green("  latency:\n    mean: %v\n    sd: %v (%.2f%%)\n", time.Duration(lMean), time.Duration(lSd), lRsd)

	mMean, mSd := sd(cnts)
	mRsd := mSd * 100 / mMean
	color.Green("  messages:\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", mMean, mSd, mRsd)

	bdmMean, bdmSd := sd(bdMergeds)
	bdmRsd := bdmSd * 100 / bdmMean
	color.Green("  bd merged (orbd2):\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", bdmMean, bdmSd, bdmRsd)

	dmMean, dmSd := sd(dMergeds)
	dmRsd := dmSd * 100 / dmMean
	color.Green("  d merged (ord5):\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", dmMean, dmSd, dmRsd)

	pmMean, pmSd := sd(pMergeds)
	pmRsd := pmSd * 100 / pmMean
	color.Green("  payloads merged (ord6):\n    mean: %.2f\n    sd: %.2f (%.2f%%)\n", pmMean, pmSd, pmRsd)

	tMean, tSd := sd(transmits)
	tRsd := tSd * 100 / tMean
	color.Green("  transmits:\n    mean: %.2f (~%.2f per broadcast)\n    sd: %.2f (%.2f%%)\n", tMean,
		tMean/float64(messages), tSd, tRsd)

	color.Blue("config:")
	color.Blue("  nodes: %v\n  connectivity (k): %v\n  byzantine nodes (f): %v"+
		"\n  runs: %v\n  protocol: %v\n  payload size: %v bytes\n  messages broadcasted: %v\n",
		runCfg.N, runCfg.K, runCfg.F, runCfg.Runs, reflect.TypeOf(runCfg.Protocol).Elem().Name(), runCfg.PayloadSize, messages)

	ctl.FlushProcesses()
	ctl.Close()

	color.Black("==========")
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
