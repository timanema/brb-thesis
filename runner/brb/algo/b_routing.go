package algo

import (
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"rp-runner/graphs"
)

type BrachaInclusionTable map[uint64][]uint64

func FindBrachaInclusionTable(g *simple.WeightedUndirectedGraph, nodes []uint64, n, f int) BrachaInclusionTable {
	res := make(BrachaInclusionTable)

	echoReq := int(math.Ceil((float64(n)+float64(f)+1)/2)) + f

	for _, n := range nodes {
		res[n] = graphs.BroadcastCostEstimation(g, n, echoReq)
	}

	return res
}
