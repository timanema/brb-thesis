package graphs

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
	paths, err := DisjointPaths(Directed(gr), a, d, k, nil, false)

	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.True(t, VerifySolution(Directed(gr), a, d, k, paths))
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
	paths, err := DisjointPaths(Directed(gr), a, d, k, nil, false)

	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.True(t, VerifySolution(Directed(gr), a, d, k, paths))
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
	_, err := DisjointPaths(Directed(gr), a, d, k, nil, false)
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

		paths, err := DisjointPaths(Directed(g), s, t, f, nil, false)
		assert.NoError(test, err)
		assert.Len(test, paths, f)
		assert.True(test, VerifySolution(Directed(g), s, t, f, paths))
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

		paths, err := DisjointPaths(Directed(g), s, t, f, nil, false)
		assert.NoError(test, err)
		assert.Len(test, paths, f)
		assert.True(test, VerifySolution(Directed(g), s, t, f, paths))
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
		p, _ = DisjointPaths(gd, s, t, k, nil, false)
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

var table map[uint64][]Path

func benchTable(n, k int, m Generator, b *testing.B) {
	g, err := m.Generate(n, k)
	if err != nil {
		b.Fail()
		return
	}

	start := rand.Intn(n)

	s := g.Node(int64(start))
	var res map[uint64][]Path

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, _ = BuildLookupTable(g, s, k, 0, false)
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

Everything above + SLICES EVERYWHERE:
BenchmarkWithGeneralizedWheelGenerator10Nodes3Connected-12      	   10000	    100631 ns/op
BenchmarkWithGeneralizedWheelGenerator10Nodes5Connected-12      	    8479	    138468 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes8Connected-12      	    1557	    688879 ns/op
BenchmarkWithGeneralizedWheelGenerator30Nodes15Connected-12     	    1012	   1174852 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes15Connected-12     	     500	   2235840 ns/op
BenchmarkWithGeneralizedWheelGenerator50Nodes25Connected-12     	     344	   3468295 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes25Connected-12    	     132	   8702465 ns/op
BenchmarkWithGeneralizedWheelGenerator100Nodes50Connected-12    	      64	  17659181 ns/op
BenchmarkTable10Nodes3Connected-12                              	    1268	    930273 ns/op
BenchmarkTable10Nodes5Connected-12                              	     942	   1198701 ns/op
BenchmarkTable30Nodes8Connected-12                              	      55	  21920535 ns/op
BenchmarkTable30Nodes15Connected-12                             	      33	  35263790 ns/op
BenchmarkTable50Nodes15Connected-12                             	       9	 121143846 ns/op
BenchmarkTable50Nodes25Connected-12                             	       6	 179129677 ns/op
BenchmarkTable100Nodes25Connected-12                            	       2	 947767060 ns/op
BenchmarkTable100Nodes50Connected-12                            	       1	1760987702 ns/op

*/

/*
Benches for presentation
*/

// 143493 ns/op == 0.14ms
func BenchmarkSingle10Nodes3ConnectedPres(b *testing.B) {
	n, k := 10, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 1031872 ns/op == 1ms
func BenchmarkSingle50Nodes3ConnectedPres(b *testing.B) {
	n, k := 50, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 2690489 ns/op == 2.7ms
func BenchmarkSingle100Nodes3ConnectedPres(b *testing.B) {
	n, k := 100, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 5166703 ns/op == 5.2ms
func BenchmarkSingle150Nodes3ConnectedPres(b *testing.B) {
	n, k := 150, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 11158237 ns/op == 11.2ms
func BenchmarkSingle200Nodes3ConnectedPres(b *testing.B) {
	n, k := 200, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 39731932 ns/op == 39.7ms
func BenchmarkSingle500Nodes3ConnectedPres(b *testing.B) {
	n, k := 500, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 191588075 ns/op == 191.6ms
func BenchmarkSingle1000Nodes3ConnectedPres(b *testing.B) {
	n, k := 1000, 3
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 5322925 ns/op == 5.3ms
func BenchmarkSingle50Nodes25ConnectedPres(b *testing.B) {
	n, k := 50, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 13095442 ns/op == 13ms
func BenchmarkSingle100Nodes25ConnectedPres(b *testing.B) {
	n, k := 100, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 20936914 ns/op == 20.9ms
func BenchmarkSingle150Nodes25ConnectedPres(b *testing.B) {
	n, k := 150, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 31597568 ns/op == 31.6ms
func BenchmarkSingle200Nodes25ConnectedPres(b *testing.B) {
	n, k := 200, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 127276025 ns/op == 127.3ms
func BenchmarkSingle500Nodes25ConnectedPres(b *testing.B) {
	n, k := 500, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 438290766 ns/op == 438.3ms
func BenchmarkSingle1000Nodes25ConnectedPres(b *testing.B) {
	n, k := 1000, 25
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 24694959 ns/op == 24.7ms
func BenchmarkSingle100Nodes50ConnectedPres(b *testing.B) {
	n, k := 100, 50
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 40792549 ns/op == 40.8ms
func BenchmarkSingle150Nodes50ConnectedPres(b *testing.B) {
	n, k := 150, 50
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 72140659 ns/op == 72.1ms
func BenchmarkSingle200Nodes50ConnectedPres(b *testing.B) {
	n, k := 200, 50
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 318277658 ns/op == 318.3ms
func BenchmarkSingle500Nodes50ConnectedPres(b *testing.B) {
	n, k := 500, 50
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

// 721954844 ns/op == 722ms
func BenchmarkSingle1000Nodes50ConnectedPres(b *testing.B) {
	n, k := 1000, 50
	benchWithGenerator(n, k, GeneralizedWheelGenerator{}, b)
}

/// TABLES
// 1791754 ns/op == 0.002s
func BenchmarkTable10Nodes5ConnectedPres(b *testing.B) {
	n, k := 10, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 45860642 ns/op == 0.05s
func BenchmarkTable50Nodes5ConnectedPres(b *testing.B) {
	n, k := 50, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 336122326 ns/op == 0.3s
func BenchmarkTable100Nodes5ConnectedPres(b *testing.B) {
	n, k := 100, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 1100778497 ns/op == 1.1s
func BenchmarkTable150Nodes5ConnectedPres(b *testing.B) {
	n, k := 150, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 2012939823 ns/op == 2s
func BenchmarkTable200Nodes5ConnectedPres(b *testing.B) {
	n, k := 200, 5
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 264568676 ns/op == 0.26s
func BenchmarkTable50Nodes25ConnectedPres(b *testing.B) {
	n, k := 50, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 1285495798 ns/op == 1.3s
func BenchmarkTable100Nodes25ConnectedPres(b *testing.B) {
	n, k := 100, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 3634130785 ns/op == 3.6s
func BenchmarkTable150Nodes25ConnectedPres(b *testing.B) {
	n, k := 150, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 7110399176 ns/op == 7.1s
func BenchmarkTable200Nodes25ConnectedPres(b *testing.B) {
	n, k := 200, 25
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 2561879285 ns/op == 2.56s
func BenchmarkTable100Nodes50ConnectedPres(b *testing.B) {
	n, k := 100, 50
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 7065331046 ns/op == 7s
func BenchmarkTable150Nodes50ConnectedPres(b *testing.B) {
	n, k := 150, 50
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}

// 13072465481 ns/op == 13.1s
func BenchmarkTable200Nodes50ConnectedPres(b *testing.B) {
	n, k := 200, 50
	benchTable(n, k, GeneralizedWheelGenerator{}, b)
}
