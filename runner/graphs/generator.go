package graphs

import (
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"strconv"
)

type Generator interface {
	Generate(n, k int) (*simple.WeightedUndirectedGraph, error)
}

func node(g *simple.WeightedUndirectedGraph, id int) graph.Node {
	n := g.Node(int64(id))

	if n == nil {
		n = Node{
			id:   int64(id),
			name: strconv.Itoa(id),
		}
		g.AddNode(n)
	}

	return n
}

// https://github.com/giovannifarina/BFT-ReliableCommunication/blob/master/multipartite_wheel.py
type MultiPartiteWheelGenerator struct{}

func (m MultiPartiteWheelGenerator) Generate(n, k int) (*simple.WeightedUndirectedGraph, error) {
	if k > n/2 || k%2 == 1 || k < 2 {
		return nil, errors.Errorf("impossible to generate (required: k < n/2, k mod 2 != 1, k > 1): n=%v, k=%v", n, k)
	}

	g := simple.NewWeightedUndirectedGraph(0, 0)

	x := n%(k/2) > 0
	numLevels := n / (k / 2)
	if x {
		numLevels += 1
	}

	levelSize := k / 2

	for i := 0; i < numLevels-1; i++ {
		for j := 0; j < levelSize; j++ {
			for z := 0; z < levelSize; z++ {
				g.SetWeightedEdge(g.NewWeightedEdge(node(g, i*levelSize+j), node(g, (i+1)*levelSize+z), 1))
			}
		}
	}

	for j := 0; j < levelSize; j++ {
		for z := 0; z < levelSize; z++ {
			g.SetWeightedEdge(g.NewWeightedEdge(node(g, (numLevels-1)*levelSize+j), node(g, z), 1))
		}
	}

	return g, nil
}

// https://github.com/giovannifarina/BFT-ReliableCommunication/blob/master/generalized_wheel.py
type GeneralizedWheelGenerator struct{}

func (w GeneralizedWheelGenerator) Generate(n, k int) (*simple.WeightedUndirectedGraph, error) {
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
