package algo

import (
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

	//FilterSubpaths(r)

	echoReq := int(math.Ceil((float64(n)+float64(f)+1)/2)) + f

	for _, nid := range nodes {
		closest := graphs.ClosestNodes(int64(nid), edges, echoReq)
		//echo[nid] = combinePaths(graphs.FilterSubpaths(buildPlan(closest, r, f)))
		echo[nid] = combinePaths(buildPlan(closest, r, f))
	}

	return echo
}
