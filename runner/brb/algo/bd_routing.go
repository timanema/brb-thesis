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
		for _, route := range r[t] {
			p := make(graphs.Path, len(route.P))
			copy(p, route.P)

			// TODO: need to set the priority for these all on true, since the deadlock solver assumes all nodes
			// deliver eventually which they don't because of partial broadcasts so might get stuck if combined with
			// partial broadcasts. Can be fixed, but reduces effectiveness of payload and dolev relay merging.
			res = append(res, Path{
				P:    p,
				Prio: true,
			})
		}
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
