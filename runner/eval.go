package main

import (
	"fmt"
	"os"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
)

func brachaDolevIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, k, fx := 75, 50, 24
	messages := 1
	deg := k
	gen := graphs.RandomRegularGenerator{}
	_, name := gen.Cache()

	cache := graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 50, 10
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 24, 11
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 24, 5
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 10, 4
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 10, 2
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaDolevFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, k, fx := 75, 60, 24
	messages := 1
	deg := k
	gen := graphs.RandomRegularGenerator{}
	_, name := gen.Cache()

	cache := graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 50, 24
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 40, 19
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 30, 14
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 20, 9
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 75, 10, 4
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.BrachaDolevKnownImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func dolevIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, k, fx := 150, 40, 10 //TODO: change to 48 for second eval run
	messages := 1
	deg := k
	gen := graphs.RandomRegularGenerator{}
	_, name := gen.Cache()

	cache := graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 100, 20
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 75, 36
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 75, 15
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 50, 24
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 50, 10
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 25, 12
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 25, 5
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 10, 4
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 10, 2
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func dolevFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, k, fx := 150, 100, 49
	messages := 1
	deg := k
	gen := graphs.RandomRegularGenerator{}
	_, name := gen.Cache()

	cache := graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 90, 44
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 80, 39
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 70, 34
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 60, 29
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 50, 24
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 40, 19
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 30, 14
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 20, 9
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, k, fx = 150, 10, 4
	cache = graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name, n, k), Gen: gen}
	if err := runMultipleMessagesTest(info, 5, n, k, fx, deg, messages, payloadSize, cache, cfg, opts, &brb.DolevKnownImprovedBD{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaIndividualTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, fx := 150, 10
	messages := 1
	gen := graphs.FullyConnectedGenerator{}

	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 150, 15
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 150, 25
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 150, 35
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 150, 49
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}

func brachaFullTests(opts brb.OptimizationConfig, info ctrl.Config, cfg process.Config, payloadSize int) {
	n, fx := 100, 25
	messages := 1
	gen := graphs.FullyConnectedGenerator{}

	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 90, 22
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 80, 20
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 70, 17
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 60, 15
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 50, 12
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 40, 10
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 30, 7
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 20, 5
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}

	n, fx = 10, 2
	if err := runMultipleMessagesTest(info, 5, n, n, fx, 0, messages, payloadSize, gen, cfg, opts, &brb.BrachaImproved{}); err != nil {
		fmt.Printf("err while running simple test: %v\n", err)
		os.Exit(1)
	}
}
