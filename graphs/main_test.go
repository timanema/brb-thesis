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
	n, k := 50, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected(b *testing.B) {
	n, k := 100, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

func BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected(b *testing.B) {
	n, k := 100, 50
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

/**
Base:
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	    3865	    345823 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    2697	    517689 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	     166	   7831245 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	      70	  15699834 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	      70	  21631389 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	      30	  37195206 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	       4	 349611368 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	       5	 220853874 ns/op
BenchmarkTable10Nodes3Connected-12                              	     457	   2575537 ns/op
BenchmarkTable10Nodes5Connected-12                              	     205	   6718679 ns/op
BenchmarkTable30Nodes8Connected-12                              	       6	 201928934 ns/op
BenchmarkTable30Nodes15Connected-12                             	       3	 357456871 ns/op
BenchmarkTable50Nodes15Connected-12                             	       2	1192750057 ns/op
BenchmarkTable50Nodes25Connected-12                             	       1	2465568932 ns/op
BenchmarkTable100Nodes25Connected-12                            	       1	19057124926 ns/op
BenchmarkTable100Nodes50Connected-12                            	       1	42178862485 ns/op

Switch to slices instead of iterators:
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	    7099	    210550 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    3826	    321511 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	     652	   2960607 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	     154	   8261997 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	      88	  17283750 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	      55	  21344657 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	       6	 177579948 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	       6	 197419463 ns/op
BenchmarkTable10Nodes3Connected-12                              	     494	   2712734 ns/op
BenchmarkTable10Nodes5Connected-12                              	     295	   4621592 ns/op
BenchmarkTable30Nodes8Connected-12                              	       9	 118539123 ns/op
BenchmarkTable30Nodes15Connected-12                             	       6	 218139893 ns/op
BenchmarkTable50Nodes15Connected-12                             	       2	 629360757 ns/op
BenchmarkTable50Nodes25Connected-12                             	       1	1615350385 ns/op
BenchmarkTable100Nodes25Connected-12                            	       1	7586956374 ns/op
BenchmarkTable100Nodes50Connected-12                            	       1	25248147767 ns/op

Reuse old distance tables (wipe connected ones):
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	    8214	    208190 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    3033	    382549 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	     472	   2714872 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	     194	   6166219 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	     100	  10849733 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	      73	  30021017 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	      15	  83804431 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	       5	 236467040 ns/op
BenchmarkTable10Nodes3Connected-12                              	    2653	    387902 ns/op
BenchmarkTable10Nodes5Connected-12                              	     282	   3605100 ns/op
BenchmarkTable30Nodes8Connected-12                              	     318	   4234642 ns/op
BenchmarkTable30Nodes15Connected-12                             	       6	 188379784 ns/op
BenchmarkTable50Nodes15Connected-12                             	      72	  20749805 ns/op
BenchmarkTable50Nodes25Connected-12                             	      49	  70094168 ns/op
BenchmarkTable100Nodes25Connected-12                            	       8	 149239317 ns/op
BenchmarkTable100Nodes50Connected-12                            	       1	20393458356 ns/op

Reuse old distance tables (wipe remembered ones):
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	    7918	    209777 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    2955	    415159 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	     432	   2429863 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	     171	   6580896 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	     111	  11104082 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	      60	  24735742 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	      22	  63510129 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	       6	 207831366 ns/op
BenchmarkTable10Nodes3Connected-12                              	    2932	    368655 ns/op
BenchmarkTable10Nodes5Connected-12                              	    1474	    879185 ns/op
BenchmarkTable30Nodes8Connected-12                              	     282	  73081211 ns/op
BenchmarkTable30Nodes15Connected-12                             	      91	 180891976 ns/op
BenchmarkTable50Nodes15Connected-12                             	       2	 599728084 ns/op
BenchmarkTable50Nodes25Connected-12                             	       1	1399980135 ns/op
BenchmarkTable100Nodes25Connected-12                            	       8	5770179206 ns/op
BenchmarkTable100Nodes50Connected-12                            	       7	18469345961 ns/op

Reuse, wipe, and switch to slices instead of maps:
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	    9501	    133033 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    7537	    359492 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	     854	   1683172 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	     291	   4411809 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	     157	   7502484 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	      63	  17593603 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	      48	  38515104 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	       9	 126216370 ns/op
BenchmarkTable10Nodes3Connected-12                              	    5394	    218768 ns/op
BenchmarkTable10Nodes5Connected-12                              	     686	   2042931 ns/op
BenchmarkTable30Nodes8Connected-12                              	     452	   3544022 ns/op
BenchmarkTable30Nodes15Connected-12                             	     111	 132636225 ns/op
BenchmarkTable50Nodes15Connected-12                             	      64	  27034601 ns/op
BenchmarkTable50Nodes25Connected-12                             	      24	 890669826 ns/op
BenchmarkTable100Nodes25Connected-12                            	      24	  77603480 ns/op
BenchmarkTable100Nodes50Connected-12                            	       1	12191080849 ns/op
*/
