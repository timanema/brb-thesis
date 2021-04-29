package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type Node struct {
	id       int64
	name     string
	original graph.Node
}

// ID returns the ID number of the node.
func (n Node) ID() int64 {
	return n.id
}

func (n Node) String() string {
	return n.name
}

func NewNodeUndirected(g *simple.WeightedUndirectedGraph, name string) graph.Node {
	return Node{
		id:   g.NewNode().ID(),
		name: name,
	}
}

func NewNodeSplit(g *simple.WeightedDirectedGraph, name string, original graph.Node) graph.Node {
	return Node{
		id:       g.NewNode().ID(),
		name:     name,
		original: original,
	}
}

type Path []graph.WeightedEdge
