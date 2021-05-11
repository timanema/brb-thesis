package graphs

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type Node struct {
	Id       int64
	Name     string
	original graph.Node
}

// ID returns the ID number of the node.
func (n Node) ID() int64 {
	return n.Id
}

func (n Node) String() string {
	return n.Name
}

func NewNodeUndirected(g *simple.WeightedUndirectedGraph, name string) graph.Node {
	return Node{
		Id:   g.NewNode().ID(),
		Name: name,
	}
}

func NewNodeSplit(g *simple.WeightedDirectedGraph, name string, original graph.Node) graph.Node {
	return Node{
		Id:       g.NewNode().ID(),
		Name:     name,
		original: original,
	}
}

type Path []graph.WeightedEdge
