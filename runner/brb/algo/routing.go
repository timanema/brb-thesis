package algo

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/graphs"
	"strconv"
)

func Routing(routes RoutingTable, id uint64, g *simple.WeightedUndirectedGraph, w, n, f int, singleHopNeighbour, combineNext, filterSubpath, bd bool) (BroadcastPlan, BrachaDolevRoutingTable) {
	if routes == nil {
		var err error
		routes, err = BuildRoutingTable(g, graphs.Node{
			Id:   int64(id),
			Name: strconv.Itoa(int(id)),
		}, 2*f+1, w, singleHopNeighbour)
		if err != nil {
			panic(fmt.Sprintf("process %v errored while building lookup table: %v\n", id, err))
		}
	}

	FixDeadlocks(routes)

	var bdPlan BrachaDolevRoutingTable
	if bd {
		nodes, m := graphs.Nodes(g)
		bdPlan = BrachaDolevRouting(routes, graphs.FindAdjMap(graphs.Directed(g), m), nodes, n, f)
	}

	return DolevRouting(routes, combineNext, filterSubpath), bdPlan
}
