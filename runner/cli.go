package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"rp-runner/brb"
	"rp-runner/ctrl"
	"rp-runner/graphs"
	"rp-runner/process"
	"strings"
	"time"
)

type EnumValue struct {
	Enum     []string
	Default  string
	selected string
}

func (e *EnumValue) Set(value string) error {
	for _, enum := range e.Enum {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", strings.Join(e.Enum, ", "))
}

func (e EnumValue) String() string {
	if e.selected == "" {
		return e.Default
	}
	return e.selected
}

func CLI() {
	app := &cli.App{
		Name:                 "rp-runner",
		Usage:                "CLI tool for Bachelor Thesis project Tim Anema",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name: "run",
				Flags: []cli.Flag{
					&cli.GenericFlag{
						Name: "template",
						Value: &EnumValue{
							Enum: []string{"brachaDolevIndividualTests", "brachaDolevFullTests",
								"brachaDolevScaleTests", "dolevIndividualTests", "dolevFullTests",
								"dolevScaleTests", "brachaIndividualTests", "brachaFullTests", "brachaScaleTests"},
							Default: "",
						},
						Usage: "select the template to use: brachaDolevIndividualTests | brachaDolevFullTests |" +
							" brachaDolevScaleTests | dolevIndividualTests | dolevFullTests | dolevScaleTests |" +
							" brachaIndividualTests | brachaFullTests | brachaScaleTests",
					},
					&cli.GenericFlag{
						Name:    "protocol",
						Aliases: []string{"p"},
						Value: &EnumValue{
							Enum:    []string{"dolev", "bracha", "brachaDolev"},
							Default: "dolev",
						},
						Usage: "select the template to use: dolev | bracha | brachaDolev (default: dolev)",
					},
					&cli.GenericFlag{
						Name:    "generator",
						Aliases: []string{"gen"},
						Value: &EnumValue{
							Enum:    []string{"randomRegular", "multiPartite", "fullyConnected", "generalizedWheel"},
							Default: "randomRegular",
						},
						Usage: "select the template to use: randomRegular | multiPartite |" +
							" fullyConnected | generalizedWheel (default: randomRegular)",
					},
					&cli.IntFlag{
						Name:  "skip",
						Usage: "set the amount of template tests to skip",
						Value: 0,
					},
					&cli.IntFlag{
						Name:  "runs",
						Usage: "set the amount of times to run tests",
						Value: 5,
					},
					&cli.IntFlag{
						Name:    "nodes",
						Aliases: []string{"n"},
						Usage:   "amount of nodes",
						Value:   25,
					},
					&cli.IntFlag{
						Name:    "connectivity",
						Aliases: []string{"k"},
						Usage:   "network connectivity",
						Value:   8,
					},
					&cli.IntFlag{
						Name:        "degree",
						Aliases:     []string{"deg"},
						DefaultText: "k",
						Value:       -1,
						Usage:       "network connectivity (degree)",
					},
					&cli.IntFlag{
						Name:    "byzantine",
						Aliases: []string{"f"},
						Usage:   "amount of byzantine nodes",
						Value:   3,
					},
					&cli.IntFlag{
						Name:    "payload",
						Aliases: []string{"ps"},
						Usage:   "payload size (in bytes)",
						Value:   12,
					},
					&cli.IntFlag{
						Name:    "verbosity",
						Aliases: []string{"v"},
						Usage:   "set verbosity (0, 1, 2, 3)",
						Value:   1,
					},
					&cli.BoolFlag{
						Name:  "multiple",
						Usage: "enable the use of multiple (N-F) transmitters",
					},
					&cli.BoolFlag{
						Name:  "cache",
						Usage: "use graph cache",
					},

					&cli.BoolFlag{
						Name:  "ord1",
						Usage: "enable ord1 (filtering of subpaths)",
					},
					&cli.BoolFlag{
						Name:  "ord2",
						Usage: "enable ord2 (single hop to neighbours)",
					},
					&cli.BoolFlag{
						Name:  "ord3",
						Usage: "enable ord3 (next hop merge)",
					},
					&cli.BoolFlag{
						Name:  "ord4",
						Usage: "enable ord4 (path reuse)",
					},
					&cli.BoolFlag{
						Name:  "ord5",
						Usage: "enable ord5 (relay merging)",
					},
					&cli.BoolFlag{
						Name:  "ord6",
						Usage: "enable ord6 (payload merging)",
					},
					&cli.BoolFlag{
						Name:  "ord7",
						Usage: "enable ord7 (implicit paths)",
					},

					&cli.BoolFlag{
						Name:  "orb1",
						Usage: "enable orb1 (implicit echo)",
					},
					&cli.BoolFlag{
						Name:  "orb2",
						Usage: "enable orb2 (minimal subset)",
					},

					&cli.BoolFlag{
						Name:  "orbd1",
						Usage: "enable orbd1 (partial broadcast)",
					},
					&cli.BoolFlag{
						Name:  "orbd2",
						Usage: "enable orbd2 (bracha dolev merge)",
					},

					&cli.BoolFlag{
						Name:  "no-color",
						Usage: "disable color printing to console",
					},
				},
				Usage: "Run a test",
				Action: func(c *cli.Context) error {
					if c.Int("verbosity") > 3 {
						_ = c.Set("verbosity", "3")
					} else if c.Int("verbosity") < 0 {
						_ = c.Set("verbosity", "0")
					}

					if c.Bool("no-color") {
						color.NoColor = true
					}

					if c.IsSet("template") {
						return runTemplate(c)
					} else {
						return runSingle(c)
					}
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runTemplate(c *cli.Context) error {
	info := ctrl.Config{
		PollDelay:  time.Millisecond * 200,
		CtrlBuffer: 2000,
		ProcBuffer: 50000,
		Verbosity:  ctrl.Verbosity(c.Int("verbosity")),
	}
	cfg := process.Config{
		MaxRetries:     5,
		RetryDelay:     time.Millisecond * 100,
		NeighbourDelay: time.Millisecond * 300,
	}

	// Optimizations
	opts := brb.OptimizationConfig{
		DolevFilterSubpaths:         c.Bool("ord1"),
		DolevSingleHopNeighbour:     c.Bool("ord2"),
		DolevCombineNextHops:        c.Bool("ord3"),
		DolevReusePaths:             c.Bool("ord4"),
		DolevRelayMerging:           c.Bool("ord5"),
		DolevPayloadMerging:         c.Bool("ord6"),
		DolevImplicitPath:           c.Bool("ord7"),
		BrachaImplicitEcho:          c.Bool("orb1"),
		BrachaMinimalSubset:         c.Bool("orb2"),
		BrachaDolevPartialBroadcast: c.Bool("orbd1"),
		BrachaDolevMerge:            c.Bool("orbd2"),
	}
	payloadSize := c.Int("payload")
	runs := c.Int("runs")
	multiple := c.Bool("multiple")

	template := c.Generic("template").(*EnumValue).selected

	color.Cyan("running template: %v\noptimizations: %+v\n\n", template, opts)

	skip := c.Int("skip")
	if skip > 0 {
		color.Yellow("skipping %v tests\n", skip)
	}

	switch template {
	case "brachaDolevIndividualTests":
		brachaDolevIndividualTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "brachaDolevFullTests":
		brachaDolevFullTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "dolevScaleTests":
		dolevScaleTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "brachaDolevScaleTests":
		brachaDolevScaleTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "dolevIndividualTests":
		dolevIndividualTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "dolevFullTests":
		dolevFullTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "brachaIndividualTests":
		brachaIndividualTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "brachaFullTests":
		brachaFullTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	case "brachaScaleTests":
		brachaScaleTests(opts, info, cfg, payloadSize, runs, multiple, skip)
	}
	return nil
}

func runSingle(c *cli.Context) error {
	info := ctrl.Config{
		PollDelay:  time.Millisecond * 200,
		CtrlBuffer: 2000,
		ProcBuffer: 50000,
		Verbosity:  ctrl.Verbosity(c.Int("verbosity")),
	}
	cfg := process.Config{
		MaxRetries:     5,
		RetryDelay:     time.Millisecond * 100,
		NeighbourDelay: time.Millisecond * 300,
	}

	// Optimizations
	opts := brb.OptimizationConfig{
		DolevFilterSubpaths:         c.Bool("ord1"),
		DolevSingleHopNeighbour:     c.Bool("ord2"),
		DolevCombineNextHops:        c.Bool("ord3"),
		DolevReusePaths:             c.Bool("ord4"),
		DolevRelayMerging:           c.Bool("ord5"),
		DolevPayloadMerging:         c.Bool("ord6"),
		DolevImplicitPath:           c.Bool("ord7"),
		BrachaImplicitEcho:          c.Bool("orb1"),
		BrachaMinimalSubset:         c.Bool("orb2"),
		BrachaDolevPartialBroadcast: c.Bool("orbd1"),
		BrachaDolevMerge:            c.Bool("orbd2"),
	}

	var gen graphs.Generator
	switch c.Generic("generator").(*EnumValue).selected {
	case "multiPartite":
		gen = &graphs.MultiPartiteWheelGenerator{}
	case "fullyConnected":
		gen = &graphs.FullyConnectedGenerator{}
	case "generalizedWheel":
		gen = &graphs.GeneralizedWheelGenerator{}
	default:
		gen = &graphs.RandomRegularGenerator{}
	}

	var br brb.Protocol
	switch c.Generic("protocol").(*EnumValue).selected {
	case "bracha":
		br = &brb.BrachaImproved{}
	case "brachaDolev":
		br = &brb.BrachaDolevKnownImproved{}
	default:
		br = &brb.DolevKnownImproved{}
	}

	if c.Bool("cache") {
		_, name := gen.Cache()
		gen = &graphs.FileCacheGenerator{Name: fmt.Sprintf("generated/%v-%v-%v.graph", name,
			c.Int("nodes"), c.Int("connectivity")), Gen: gen}
	}

	runCfg := RunConfig{
		Runs:                 c.Int("runs"),
		N:                    c.Int("nodes"),
		K:                    c.Int("connectivity"),
		F:                    c.Int("byzantine"),
		Degree:               c.Int("degree"),
		PayloadSize:          c.Int("payload"),
		MultipleTransmitters: c.Bool("multiple"),
		Generator:            gen,
		ControlCfg:           info,
		ProcessCfg:           cfg,
		OptimizationCfg:      opts,
		Protocol:             br,
	}
	color.Cyan("running single run: %+v\noptimizations: %+v\n\n", runCfg, opts)

	return runMultipleMessagesTest(runCfg, false)
}
