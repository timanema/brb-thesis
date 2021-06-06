package graphs

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"unsafe"
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

func NewNodeSplit(name string, original graph.Node, in bool, n int64) graph.Node {
	if in {
		return Node{
			Id:       original.ID(),
			Name:     name,
			original: original,
		}
	}

	return Node{
		Id:       original.ID() + n,
		Name:     name,
		original: original,
	}
}

type Path []graph.WeightedEdge

func (p Path) SizeOf() uintptr {
	if len(p) == 0 {
		return 0
	}

	y := unsafe.Sizeof(p[0].(simple.WeightedEdge))
	return uintptr(len(p)) * y
}
