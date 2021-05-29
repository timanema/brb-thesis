package graphs

import (
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
)

// https://github.com/giovannifarina/BFT-ReliableCommunication/blob/master/multipartite_wheel.py
type MultiPartiteWheelGenerator struct{}

func (m MultiPartiteWheelGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
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

func (m MultiPartiteWheelGenerator) Cache() (bool, string) {
	return true, "multipart_wheel"
}

// https://github.com/jdecouchant/BRB-partially-connected-networks/blob/c1980ed20d6a11f4230299b2887bbabd220a84c0/BroadcastSign/simulations/generateGraphs.py#L56
type MultiPartiteWheelAltGenerator struct{}

func (m MultiPartiteWheelAltGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	if k > n/2 || k%2 == 1 || k < 2 {
		return nil, errors.Errorf("impossible to generate (required: k < n/2, k mod 2 != 1, k > 1): n=%v, k=%v", n, k)
	}

	g := simple.NewWeightedUndirectedGraph(0, 0)

	// Weird rounding stuff here, idk how python handles it so just a 1:1 copy including all weird casts

	gSize := float64(k) / 2
	for gId := 0; gId < int(float64(n)/gSize)-1; gId++ {
		for nid1 := int(float64(gId) * gSize); nid1 < int(float64(gId)*gSize+gSize); nid1++ {
			for nid2 := int(float64(gId+1) * gSize); nid2 < int(float64(gId+1)*gSize+gSize); nid2++ {
				g.SetWeightedEdge(g.NewWeightedEdge(node(g, nid1), node(g, nid2), 1))
			}
		}
	}

	gId := float64(n)/gSize - 1
	for nid1 := 0; nid1 < int(gSize); nid1++ {
		for nid2 := int(gId * gSize); nid2 < int(gId*gSize+gSize); nid2++ {
			g.SetWeightedEdge(g.NewWeightedEdge(node(g, nid1), node(g, nid2), 1))
		}
	}

	if n%int(gSize) != 0 {
		for i := int(float64(n)/gSize) * int(gSize); i < n; i++ {
			for j := 0; j < k; j++ {
				g.SetWeightedEdge(g.NewWeightedEdge(node(g, i), node(g, j), 1))
			}
		}
	}

	return g, nil
}

func (m MultiPartiteWheelAltGenerator) Cache() (bool, string) {
	return true, "multipart_wheel_alt"
}
