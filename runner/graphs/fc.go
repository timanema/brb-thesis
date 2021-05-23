package graphs

import (
	"gonum.org/v1/gonum/graph/simple"
	"sort"
)

func IsFullyConnected(g *simple.WeightedUndirectedGraph) bool {
	nodes := g.Nodes()
	l := nodes.Len()

	for nodes.Next() {
		n := nodes.Node()

		o := g.From(n.ID())
		if o.Len() < l-1 {
			return false
		}
	}

	return true
}

func BroadcastCostEstimation(g *simple.WeightedUndirectedGraph, s uint64, amount int) []uint64 {
	costs := make([]cost, 0, g.Nodes().Len())
	res := make([]uint64, 0, amount)

	nodes := g.Nodes()

	if cap(res) > cap(costs) {
		panic("requesting too many nodes for broadcast cost estimation!")
	}

	for nodes.Next() {
		n := nodes.Node()

		// Skip source
		if n.ID() == int64(s) {
			continue
		}

		c := 0
		f := g.From(n.ID())
		for f.Next() {
			t := f.Node()

			c += int(g.WeightedEdge(n.ID(), t.ID()).Weight())
		}

		costs = append(costs, cost{
			id:   uint64(n.ID()),
			cost: c,
		})
	}

	// Sort costs
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].cost < costs[j].cost
	})

	res = append(res, s)
	for i := 0; i < amount-1; i++ {
		res = append(res, costs[i].id)
	}

	return res
}
