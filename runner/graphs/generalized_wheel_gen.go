package graphs

import (
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
)

// https://github.com/giovannifarina/BFT-ReliableCommunication/blob/master/generalized_wheel.py
type GeneralizedWheelGenerator struct{}

func (w GeneralizedWheelGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	if k >= n || k < 2 {
		return nil, errors.Errorf("impossible to generate (required: k < n, k > 1): n=%v, k=%v", n, k)
	}

	g := simple.NewWeightedUndirectedGraph(0, 0)
	clique := make([]int, 0)

	if k == 3 {
		clique = append(clique, 0)
		node(g, 0)
	} else {
		for i := 0; i < k-2; i++ {
			clique = append(clique, i)
			for j := 0; j < k-2; j++ {
				if i != j {
					g.SetWeightedEdge(g.NewWeightedEdge(node(g, i), node(g, j), 1))
				}
			}
		}
	}

	for i := k - 2; i < n-1; i++ {
		g.SetWeightedEdge(g.NewWeightedEdge(node(g, i), node(g, i+1), 1))

		for _, e := range clique {
			g.SetWeightedEdge(g.NewWeightedEdge(node(g, e), node(g, i), 1))
		}
	}

	g.SetWeightedEdge(g.NewWeightedEdge(node(g, k-2), node(g, n-1), 1))
	for _, e := range clique {
		g.SetWeightedEdge(g.NewWeightedEdge(node(g, e), node(g, n-1), 1))
	}

	return g, nil
}

func (w GeneralizedWheelGenerator) Cache() (bool, string) {
	return false, "gen_wheel"
}
