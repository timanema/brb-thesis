package graphs

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"math"
)

type pair struct {
	a, b int64
}

func findPairs(xs []int64) []pair {
	res := make([]pair, 0, len(xs)*len(xs))

	for i := 0; i < len(xs); i++ {
		for j := 0; j < len(xs); j++ {
			if i == j {
				continue
			}

			res = append(res, pair{a: xs[i], b: xs[j]})
		}
	}

	return res
}

func localConnectivity(split *SplitGraph, edges [][]graph.WeightedEdge, s, t graph.Node) int {
	duplicate := make([][]graph.WeightedEdge, len(edges))
	for i := range edges {
		duplicate[i] = make([]graph.WeightedEdge, len(edges[i]))
		copy(duplicate[i], edges[i])
	}

	return maxFlow(duplicate, s.ID()+int64(split.g.Nodes().Len()/2), t.ID())
}

// Based on algorithm 11 in http://www.cse.msu.edu/~cse835/Papers/Graph_connectivity_revised.pdf
func FindConnectedness(gu *simple.WeightedUndirectedGraph) int {
	g := Directed(gu)
	split := NodeSplitting(g)
	PrepareFlow(split.g, false)
	edges := FindAdjMap(split.g, int64(split.g.Nodes().Len()))

	// Find node with minimum degree
	nodes := g.Nodes()
	minV := math.MaxInt64
	var minN graph.Node

	if nodes.Len() < 2 {
		return 0
	}

	for nodes.Next() {
		n := nodes.Node()
		deg := g.From(n.ID()).Len()

		if deg < minV {
			minV = deg
			minN = n
		}
	}

	nodes.Reset()

	if minN == nil {
		return 0
	}

	neighbours := make([]int64, 0, nodes.Len())

	// Compute local connectivity with all non-neighbours
	for nodes.Next() {
		n := nodes.Node()

		if n.ID() == minN.ID() {
			continue
		}

		if g.HasEdgeFromTo(minN.ID(), n.ID()) {
			neighbours = append(neighbours, n.ID())
			continue
		}

		con := localConnectivity(split, edges, minN, n)
		if con < minV {
			minV = con
		}
	}

	// Compute local connectivity for all pairs of neighbours (non-adjacent)
	pairs := findPairs(neighbours)

	for _, pair := range pairs {
		if g.HasEdgeBetween(pair.a, pair.b) {
			continue
		}

		con := localConnectivity(split, edges, g.Node(pair.a), g.Node(pair.b))
		if con < minV {
			minV = con
		}
	}

	return minV
}
