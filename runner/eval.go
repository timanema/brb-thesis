package main

import (
	"fmt"
	"os"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
)

func brachaDolevIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		N:                    75,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaDolevKnownImproved{},
	}

	_, name := gen.Cache()
	runCfg.K = 50
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 50
	runCfg.F = 10
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 24
	runCfg.F = 11
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 24
	runCfg.F = 5
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 4
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 2
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaDolevFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		N:                    75,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaDolevKnownImproved{},
	}

	_, name := gen.Cache()
	runCfg.K = 60
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 50
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 40
	runCfg.F = 19
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 30
	runCfg.F = 14
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 20
	runCfg.F = 9
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 4
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaDolevScaleTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaDolevKnownImproved{},
	}

	_, name := gen.Cache()
	runCfg.N = 75
	runCfg.K = 24
	runCfg.F = 11
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 50
	runCfg.K = 16
	runCfg.F = 7
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 25
	runCfg.K = 8
	runCfg.F = 3
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func dolevIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		N:                    150,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.DolevKnownImprovedPM{},
	}

	_, name := gen.Cache()
	runCfg.K = 100
	runCfg.F = 48
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 100
	runCfg.F = 20
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 75
	runCfg.F = 36
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 75
	runCfg.F = 15
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 50
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 50
	runCfg.F = 10
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 25
	runCfg.F = 12
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 25
	runCfg.F = 5
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 4
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 2
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func dolevFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		N:                    150,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.DolevKnownImprovedPM{},
	}

	_, name := gen.Cache()
	runCfg.K = 100
	runCfg.F = 49
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 90
	runCfg.F = 44
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 80
	runCfg.F = 39
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 70
	runCfg.F = 34
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 60
	runCfg.F = 29
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 50
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 40
	runCfg.F = 19
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 30
	runCfg.F = 14
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 20
	runCfg.F = 9
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.K = 10
	runCfg.F = 4
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func dolevScaleTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	gen := &graphs.RandomRegularGenerator{}
	runCfg := RunConfig{
		Runs:                 runs,
		Degree:               -1,
		MultipleTransmitters: multiple,
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.DolevKnownImprovedPM{},
	}

	_, name := gen.Cache()
	runCfg.N = 150
	runCfg.K = 50
	runCfg.F = 24
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 125
	runCfg.K = 40
	runCfg.F = 19
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 100
	runCfg.K = 33
	runCfg.F = 16
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 75
	runCfg.K = 24
	runCfg.F = 11
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 50
	runCfg.K = 16
	runCfg.F = 7
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	_, name = gen.Cache()
	runCfg.N = 25
	runCfg.K = 8
	runCfg.F = 3
	runCfg.Generator = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, runCfg.N, runCfg.K), Gen: gen}
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	runCfg := RunConfig{
		Runs:                 runs,
		N:                    150,
		Degree:               -1,
		MultipleTransmitters: multiple,
		Generator:            &graphs.FullyConnectedGenerator{},
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaImproved{},
	}

	runCfg.F = 10
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.F = 15
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.F = 25
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.F = 35
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.F = 49
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	runCfg := RunConfig{
		Runs:                 runs,
		Degree:               -1,
		MultipleTransmitters: multiple,
		Generator:            &graphs.FullyConnectedGenerator{},
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaImproved{},
	}

	runCfg.N = 100
	runCfg.F = 25
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 90
	runCfg.F = 22
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 80
	runCfg.F = 20
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 70
	runCfg.F = 17
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 60
	runCfg.F = 15
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 50
	runCfg.F = 12
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 40
	runCfg.F = 10
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 30
	runCfg.F = 7
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 20
	runCfg.F = 5
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 10
	runCfg.F = 2
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaScaleTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize, runs int, multiple bool) {
	runCfg := RunConfig{
		Runs:                 runs,
		Degree:               -1,
		MultipleTransmitters: multiple,
		Generator:            &graphs.FullyConnectedGenerator{},
		PayloadSize:          payloadSize,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             &brb.BrachaImproved{},
	}

	runCfg.N = 150
	runCfg.F = 37
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 125
	runCfg.F = 31
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 100
	runCfg.F = 25
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 75
	runCfg.F = 18
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 50
	runCfg.F = 12
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	runCfg.N = 25
	runCfg.F = 6
	if err := runMultipleMessagesTest(runCfg); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}
