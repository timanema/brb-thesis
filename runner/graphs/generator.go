package graphs

import (
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"strconv"
)

type Generator interface {
	// d = degree, will only be used by random regular
	Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error)

	Cache() (bool, string)
}

func node(g *simple.WeightedUndirectedGraph, id int) graph.Node {
	n := g.Node(int64(id))

	if n == nil {
		n = Node{
			Id:   int64(id),
			Name: strconv.Itoa(id),
		}
		g.AddNode(n)
	}

	return n
}

type BasicGenerator struct {
	G *simple.WeightedUndirectedGraph
}

func (r BasicGenerator) Generate(_, _, _ int) (*simple.WeightedUndirectedGraph, error) {
	return r.G, nil
}

func (r BasicGenerator) Cache() (bool, string) {
	return false, ""
}

type FileCacheGenerator struct {
	Gen  Generator
	Name string
}

func (fc FileCacheGenerator) dump(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	g, err := fc.Gen.Generate(n, k, d)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate graph before caching in files")
	}

	if d, _ := fc.Gen.Cache(); d {
		fmt.Printf("graph %v does not exist in storage, dumping the generated graph...\n", fc.Name)

		if err := DumpToFile(g, fc.Name); err != nil {
			return nil, errors.Wrap(err, "unable to save file")
		}
	}

	return g, nil
}

func (fc FileCacheGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	if _, err := os.Stat(fc.Name); os.IsNotExist(err) {
		// Need to generate the graph first once
		fmt.Printf("graph %v does not exist in storage, generating it first...\n", fc.Name)
		return fc.dump(n, k, d)
	}

	return ReadFromFile(fc.Name)
}

func (fc FileCacheGenerator) Cache() (bool, string) {
	return false, ""
}
