package algo

import (
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"rp-runner/graphs"
)

type BrachaDolevRoutingTable map[uint64]BroadcastPlan

func buildPlan(to []uint64, r RoutingTable, f int) []Path {
	res := make([]Path, 0, 2*f+1)

	for _, t := range to {
		res = append(res, r[t]...)
	}

	return res
}

func BrachaDolevRouting(r RoutingTable, edges graphs.AdjacencyMap, nodes []uint64, n, f int) BrachaDolevRoutingTable {
	echo := make(map[uint64]BroadcastPlan)

	echoReq := int(math.Ceil((float64(n)+float64(f)+1)/2)) + f

	for _, nid := range nodes {
		closest := graphs.ClosestNodes(int64(nid), edges, echoReq)
		echo[nid] = combinePaths(buildPlan(closest, r, f))
	}

	return echo
}

func FindBrachaDolevInclusionTable(g *simple.WeightedUndirectedGraph, nodes []uint64, n, f int) BrachaInclusionTable {
	res := make(BrachaInclusionTable)

	echoReq := int(math.Ceil((float64(n)+float64(f)+1)/2)) + f
	edges := graphs.FindAdjMap(graphs.Directed(g), graphs.MaxId(g))

	for _, n := range nodes {
		res[n] = graphs.ClosestNodes(int64(n), edges, echoReq)
	}

	return res
}
