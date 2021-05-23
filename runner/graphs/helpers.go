package graphs

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func EqualEdge(a, b graph.Edge) bool {
	return a.To().ID() == b.To().ID() && a.From().ID() == b.From().ID()
}

func IsSubPath(a, b Path) bool {
	if len(a) > len(b) {
		return false
	}

	for i, e := range a {
		if !EqualEdge(e, b[i]) {
			return false
		}
	}

	return true
}

func FilterSubpaths(paths []Path) []Path {
	res := make([]Path, 0, len(paths))

	for i, p := range paths {
		sub := false
		for j, sp := range paths {
			if i == j {
				continue
			}

			if IsSubPath(p, sp) {
				sub = true
				break
			}
		}

		if !sub {
			res = append(res, p)
		}
	}

	return res
}

func Directed(g *simple.WeightedUndirectedGraph) *simple.WeightedDirectedGraph {
	gr := simple.NewWeightedDirectedGraph(0, 0)
	nodes := g.Nodes()

	// Copy nodes (while maintaining their IDs)
	for nodes.Next() {
		n := nodes.Node()
		gr.AddNode(n)
	}

	// Copy edges (since IDs are identical, other nodes can be used directly)
	edges := g.Edges()
	for edges.Next() {
		e := edges.Edge().(graph.WeightedEdge)

		gr.SetWeightedEdge(gr.NewWeightedEdge(e.From(), e.To(), e.Weight()))
		gr.SetWeightedEdge(gr.NewWeightedEdge(e.To(), e.From(), e.Weight()))
	}

	return gr
}

func PrepareFlow(g *simple.WeightedDirectedGraph, excludeZero bool) {
	edges := g.Edges()
	for edges.Next() {
		e := edges.Edge().(simple.WeightedEdge)

		if (excludeZero && e.Weight() > 0.1) || !excludeZero {
			e.W = 1
			g.SetWeightedEdge(e)
		}
	}
}

func MaxId(g *simple.WeightedUndirectedGraph) int64 {
	_, m := Nodes(g)

	return m
}

func Nodes(g *simple.WeightedUndirectedGraph) ([]uint64, int64) {
	m := int64(0)
	nodes := g.Nodes()
	res := make([]uint64, 0, nodes.Len())

	for nodes.Next() {
		n := nodes.Node()
		res = append(res, uint64(n.ID()))

		if id := n.ID(); id > m {
			m = id
		}
	}

	return res, m
}
