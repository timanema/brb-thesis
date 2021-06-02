package graphs

import (
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
)

type FullyConnectedGenerator struct{}

func (f FullyConnectedGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	if n != k {
		return nil, errors.Errorf("impossible to generate (required n=k): n=%v, k=%v", n, k)
	}
	g := simple.NewWeightedUndirectedGraph(0, 0)

	// Add all edges (and nodes)
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			if i == j {
				continue
			}

			g.SetWeightedEdge(g.NewWeightedEdge(node(g, i), node(g, j), 1))
		}
	}

	return g, nil
}

func (f FullyConnectedGenerator) Cache() (bool, string) {
	return false, "fc"
}
