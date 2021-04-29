package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"math/rand"
	"testing"
)

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("failed to seed random with crypto/rand")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func TestSimple(t *testing.T) {
	gr := simple.NewWeightedUndirectedGraph(0, 0)

	/*   b
	  /     \
	 0       1
	/         \
	a          d
	\         /
	 1       0
	  \     /
	     c
	*/

	// Create nodes
	a := NewNodeUndirected(gr, "a")
	gr.AddNode(a)

	b := NewNodeUndirected(gr, "b")
	gr.AddNode(b)

	c := NewNodeUndirected(gr, "c")
	gr.AddNode(c)

	d := NewNodeUndirected(gr, "d")
	gr.AddNode(d)

	// Create edges
	ab := gr.NewWeightedEdge(a, b, 0)
	gr.SetWeightedEdge(ab)

	ac := gr.NewWeightedEdge(a, c, 1)
	gr.SetWeightedEdge(ac)

	bd := gr.NewWeightedEdge(b, d, 1)
	gr.SetWeightedEdge(bd)

	cd := gr.NewWeightedEdge(c, d, 0)
	gr.SetWeightedEdge(cd)

	k := 2

	// Get paths
	paths, err := DisjointPaths(Directed(gr), a, d, k)

	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.True(t, VerifyDisjointPaths(Directed(gr), a, d, k, paths))
}

func TestSimpleTrap(t *testing.T) {
	gr := simple.NewWeightedUndirectedGraph(0, 0)

	/*    b
	   /  |  \
	  0   |   1
	 /    |    \
	a     0     d
	 \    |    /
	  1   |   0
	   \  |  /
	      c
	*/

	// Create nodes
	a := NewNodeUndirected(gr, "a")
	gr.AddNode(a)

	b := NewNodeUndirected(gr, "b")
	gr.AddNode(b)

	c := NewNodeUndirected(gr, "c")
	gr.AddNode(c)

	d := NewNodeUndirected(gr, "d")
	gr.AddNode(d)

	// Create edges
	ab := gr.NewWeightedEdge(a, b, 0)
	gr.SetWeightedEdge(ab)

	ac := gr.NewWeightedEdge(a, c, 1)
	gr.SetWeightedEdge(ac)

	bd := gr.NewWeightedEdge(b, d, 1)
	gr.SetWeightedEdge(bd)

	cd := gr.NewWeightedEdge(c, d, 0)
	gr.SetWeightedEdge(cd)

	bc := gr.NewWeightedEdge(b, c, 0)
	gr.SetWeightedEdge(bc)

	k := 2

	// Get paths
	paths, err := DisjointPaths(Directed(gr), a, d, k)

	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.True(t, VerifyDisjointPaths(Directed(gr), a, d, k, paths))
}

func TestImpossible(t *testing.T) {
	gr := simple.NewWeightedUndirectedGraph(0, 0)

	/*   b
	  /     \
	 0       1
	/         \
	a          d
	*/

	// Create nodes
	a := NewNodeUndirected(gr, "a")
	gr.AddNode(a)

	b := NewNodeUndirected(gr, "b")
	gr.AddNode(b)

	d := NewNodeUndirected(gr, "d")
	gr.AddNode(d)

	// Create edges
	ab := gr.NewWeightedEdge(a, b, 0)
	gr.SetWeightedEdge(ab)

	bd := gr.NewWeightedEdge(b, d, 1)
	gr.SetWeightedEdge(bd)

	// K is too high, impossible
	k := 2

	// Get paths
	_, err := DisjointPaths(Directed(gr), a, d, k)
	assert.Error(t, err)
}

func TestWithMultiPartiteWheelGenerator(test *testing.T) {
	m := MultiPartiteWheelGenerator{}

	for n := 4; n < 70; n++ {
		k := n / 2
		// Added check for multipartite wheel
		if k%2 == 1 {
			k -= 1
		}

		f := int(math.Min(1, float64(k-2)))

		g, err := m.Generate(n, k)
		assert.NoError(test, err)

		start := rand.Intn(n)
		end := rand.Intn(n)
		for start == end {
			end = rand.Intn(n)
		}

		s, t := g.Node(int64(start)), g.Node(int64(end))

		paths, err := DisjointPaths(Directed(g), s, t, f)
		assert.NoError(test, err)
		assert.Len(test, paths, f)
		assert.True(test, VerifyDisjointPaths(Directed(g), s, t, f, paths))
	}
}

func TestWithGeneralizedWheelGenerator(test *testing.T) {
	m := GeneralizedWheelGenerator{}

	for n := 4; n < 70; n++ {
		k := n / 2
		f := int(math.Min(1, float64(k-2)))

		g, err := m.Generate(n, k)
		assert.NoError(test, err)

		start := rand.Intn(n)
		end := rand.Intn(n)
		for start == end {
			end = rand.Intn(n)
		}

		s, t := g.Node(int64(start)), g.Node(int64(end))

		paths, err := DisjointPaths(Directed(g), s, t, f)
		assert.NoError(test, err)
		assert.Len(test, paths, f)
		assert.True(test, VerifyDisjointPaths(Directed(g), s, t, f, paths))
	}
}

var paths []Path

func benchWithGenerator(n, k int, m Generator, b *testing.B) {
	g, err := m.Generate(n, k)
	if err != nil {
		b.Fail()
		return
	}

	start := rand.Intn(n)
	end := rand.Intn(n)
	for start == end {
		end = rand.Intn(n)
	}

	s, t := g.Node(int64(start)), g.Node(int64(end))
	var p []Path

	b.ResetTimer()

	gd := Directed(g)
	for i := 0; i < b.N; i++ {
		p, _ = DisjointPaths(gd, s, t, k)
	}

	paths = p
}

func BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected(b *testing.B) {
	n, k := 10, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected(b *testing.B) {
	n, k := 10, 5
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected(b *testing.B) {
	n, k := 30, 8
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected(b *testing.B) {
	n, k := 30, 15
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected(b *testing.B) {
	n, k := 50, 15
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected(b *testing.B) {
	n, k := 50, 15
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected(b *testing.B) {
	n, k := 100, 40
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected(b *testing.B) {
	n, k := 100, 40
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

var table map[int64][]Path

func benchTable(n, k int, m Generator, b *testing.B) {
	g, err := m.Generate(n, k)
	if err != nil {
		b.Fail()
		return
	}

	start := rand.Intn(n)

	s := g.Node(int64(start))
	var res map[int64][]Path

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, _ = BuildLookupTable(g, s, k)
	}

	table = res
}

func BenchmarkTable10Nodes3Connected(b *testing.B) {
	n, k := 10, 3
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable10Nodes5Connected(b *testing.B) {
	n, k := 10, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable30Nodes8Connected(b *testing.B) {
	n, k := 30, 8
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable30Nodes15Connected(b *testing.B) {
	n, k := 30, 15
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable50Nodes15Connected(b *testing.B) {
	n, k := 50, 15
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable50Nodes25Connected(b *testing.B) {
	n, k := 50, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable100Nodes25Connected(b *testing.B) {
	n, k := 100, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkTable100Nodes50Connected(b *testing.B) {
	n, k := 100, 50
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}
