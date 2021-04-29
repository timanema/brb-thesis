package main

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"math/rand"
	"time"
)

// Pretty printing of graphs (pasted from graphviz wiki, slightly curated)
var colors = []string{
	"aquamarine",
	"bisque",
	"blue",
	"blueviolet",
	"brown",
	"burlywood",
	"cadetblue",
	"chartreuse",
	"chocolate",
	"coral",
	"cornflowerblue",
	"crimson",
	"cyan",
	"darkgoldenrod",
	"darkgreen",
	"darkkhaki",
	"darkolivegreen",
	"darkorange",
	"darkorchid",
	"darksalmon",
	"darkseagreen",
	"darkslateblue",
	"darkturquoise",
	"darkviolet",
	"deeppink",
	"deepskyblue",
	"dimgray",
	"dodgerblue",
	"firebrick",
	"firebrick",
	"forestgreen",
	"gold",
	"goldenrod",
	"grey",
	"green",
	"greenyellow",
	"hotpink",
	"indianred",
	"indigo",
	"khaki",
	"lawngreen",
	"lightblue",
	"lightcoral",
	"lightgray",
	"lightpink",
	"lightsalmon",
	"lightseagreen",
	"lightskyblue",
	"lightslategray",
	"lightsteelblue",
	"limegreen",
	"magenta",
	"maroon",
	"mediumaquamarine",
	"mediumblue",
	"mediumorchid",
	"mediumpurple",
	"mediumseagreen",
	"mediumslateblue",
	"mediumspringgreen",
	"mediumturquoise",
	"mediumvioletred",
	"midnightblue",
	"navajowhite",
	"navy",
	"olivedrab",
	"orange",
	"orangered",
	"orchid",
	"palegoldenrod",
	"palegreen",
	"paleturquoise",
	"palevioletred",
	"peachpuff",
	"peru",
	"pink",
	"plum",
	"powderblue",
	"purple",
	"red",
	"rosybrown",
	"royalblue",
	"saddlebrown",
	"salmon",
	"sandybrown",
	"seagreen",
	"sienna",
	"skyblue",
	"slateblue",
	"slategray",
	"springgreen",
	"steelblue",
	"tan",
	"thistle",
	"tomato",
	"turquoise",
	"violet",
	"wheat",
	"yellow",
	"yellowgreen",
}

func PrintGraphviz(g graph.WeightedDirected) {
	PrintGraphvizHighlightPaths(g, []Path{})
}

func PrintGraphvizHighlightPaths(g graph.WeightedDirected, paths []Path) {
	nodes := g.Nodes()
	fmt.Printf("digraph {\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	cols := make(map[int64]map[int64]string)

	for _, p := range paths {
		col := colors[r.Intn(len(colors))]

		for _, e := range p {
			if _, ok := cols[e.From().ID()]; !ok {
				cols[e.From().ID()] = make(map[int64]string)
			}

			cols[e.From().ID()][e.To().ID()] = col
		}
	}

	for nodes.Next() {
		n := nodes.Node()
		col, ok := cols[n.ID()]

		to := g.From(n.ID())
		for to.Next() {
			t := to.Node()
			w := g.WeightedEdge(n.ID(), t.ID()).Weight()

			c := "black"
			width := 1
			if ok {
				if color, ok := col[t.ID()]; ok {
					c = color
					width = 3
				}
			}

			fmt.Printf("    %v -> %v[label=\"%v\",weight=\"%v\",color=%v,penwidth=%v];\n", n, t, w, w, c, width)
		}
	}

	fmt.Printf("}\n")
}
