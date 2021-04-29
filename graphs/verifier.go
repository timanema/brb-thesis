package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func VerifyDisjointPaths(g *simple.WeightedDirectedGraph, s, t graph.Node, k int, paths []Path) bool {
	// Check if enough paths
	if len(paths) < k {
		return false
	}

	// Verify the paths are disjoint
	seen := make(map[int64]struct{})

	for _, p := range paths {
		for _, e := range p {
			// Skip start node
			if e.From().ID() == s.ID() {
				continue
			}

			// Not disjoint
			if _, ok := seen[e.From().ID()]; ok {
				return false
			}

			seen[e.From().ID()] = struct{}{}
		}
	}

	// Verify the paths are valid
	for _, p := range paths {
		// Invalid path
		if len(p) < 1 {
			return false
		}

		prev := p[0]

		// All paths should start at s
		if prev.From().ID() != s.ID() {
			return false
		}

		for i, cur := range p {
			// Every edge should be valid
			if !g.HasEdgeBetween(cur.From().ID(), cur.To().ID()) {
				return false
			}

			// Previous .To() should equal current .From() (if this is not the first edge)
			if prev.To().ID() != cur.From().ID() && i != 0 {
				return false
			}

			prev = cur
		}

		// All paths should end at t
		if prev.To().ID() != t.ID() {
			return false
		}
	}

	return true
}
