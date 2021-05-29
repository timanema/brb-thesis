package graphs

import (
	"encoding/gob"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"os"
)

type savedGraph struct {
	Nodes map[int64]struct{}
	Edges map[int64]map[int64]float64
}

func newSavedGraph(g *simple.WeightedUndirectedGraph) savedGraph {
	s := savedGraph{
		Nodes: make(map[int64]struct{}),
		Edges: make(map[int64]map[int64]float64),
	}
	nodes := g.Nodes()

	for nodes.Next() {
		n := nodes.Node()
		from := g.From(n.ID())

		s.Nodes[n.ID()] = struct{}{}
		s.Edges[n.ID()] = make(map[int64]float64)

		for from.Next() {
			to := from.Node()

			s.Edges[n.ID()][to.ID()] = g.Edge(n.ID(), to.ID()).(simple.WeightedEdge).W
		}
	}

	return s
}

func (s savedGraph) Build() *simple.WeightedUndirectedGraph {
	g := simple.NewWeightedUndirectedGraph(0, 0)

	for id, _ := range s.Nodes {
		node(g, int(id))
	}

	for f, to := range s.Edges {
		for t, w := range to {
			g.SetWeightedEdge(g.NewWeightedEdge(node(g, int(f)), node(g, int(t)), w))
		}
	}

	return g
}

func DumpToFile(g *simple.WeightedUndirectedGraph, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return errors.Wrap(err, "unable to create file")
	}
	defer f.Close()

	s := newSavedGraph(g)
	//fmt.Println(s)

	enc := gob.NewEncoder(f)
	return errors.Wrap(enc.Encode(s), "failed to encode graph")
}

func ReadFromFile(name string) (*simple.WeightedUndirectedGraph, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open file")
	}
	defer f.Close()

	var s savedGraph
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&s); err != nil {
		return nil, errors.Wrap(err, "failed to decode graph")
	}
	//fmt.Println(s)

	return s.Build(), nil
}
