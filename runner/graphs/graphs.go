package graphs

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"

	_ "net/http/pprof"
)

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("failed to seed random with crypto/rand")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func benchTableTest() {
	n, k := 30, 15
	m := GeneralizedWheelGenerator{}

	g, err := m.Generate(n, k)
	if err != nil {
		fmt.Printf("failed to generate graph for lookup test: %v\n", err)
		os.Exit(1)
	}

	start := rand.Intn(n)

	s := g.Node(int64(start))
	//PrintGraphviz(Directed(g))

	res, err := BuildLookupTable(g, s, k, 0, false)
	if err != nil {
		fmt.Printf("failed to build lookup table for %v: %v\n", start, err)
		os.Exit(1)
	}

	fmt.Println("Result:")
	for to, paths := range res {
		fmt.Printf("%v -> %v\n\n", to, paths)
	}
	//PrintGraphvizHighlightRoutes(Directed(g), res)
}

func benchSingleTest() {
	n, k := 30, 15
	m := GeneralizedWheelGenerator{}

	g, err := m.Generate(n, k)
	if err != nil {
		fmt.Printf("failed to generate graph for lookup test: %v\n", err)
		os.Exit(1)
	}

	start := rand.Intn(n)

	s := g.Node(int64(start))
	t := g.Node(int64(math.Max(1, float64(start-1))))
	f := int(math.Ceil((float64(k) - 1) / 2))
	//PrintGraphviz(Directed(g))

	additionalWeight := make(map[uint64]map[uint64]int)

	nodes := g.Nodes()
	for nodes.Next() {
		n := uint64(nodes.Node().ID())
		if _, ok := additionalWeight[n]; !ok {
			additionalWeight[n] = make(map[uint64]int)
		}
	}
	nodes.Reset()
	fmt.Printf("%v, %v\n", start, math.Max(1, float64(start-1)))
	res, err := DisjointPaths(Directed(g), nil, s, t, f, nil, false)
	if err != nil {
		fmt.Printf("failed to build single path for %v: %v\n", start, err)
		os.Exit(1)
	}

	PrintGraphvizHighlightPaths(Directed(g), res)
	fmt.Printf("Result (%s -> %s over %v paths):\n%v\n", s, t, f, res)

	_, err = DisjointPaths(Directed(g), nil, s, t, k, nil, false)
	if err != nil {
		fmt.Printf("failed to build single path for %v: %v\n", start, err)
		os.Exit(1)
	}
}

func GraphsMain() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//benchTableTest()
	//return

	nx, kx := 150, 80
	mx := GeneralizedWheelGenerator{}

	gx, errx := mx.Generate(nx, kx)
	if errx != nil {
		fmt.Printf("failed to generate graph for lookup test: %v\n", errx)
		os.Exit(1)
	}
	fmt.Println(FindConnectedness(gx))
	return

	//x := GeneralizedWheelGenerator{}
	//gx, err := x.Generate(5, 2)
	//if err != nil {
	//	fmt.Printf("invalid graph parameters: %v\n", err)
	//	os.Exit(1)
	//}
	//
	//PrintGraphviz(Directed(gx))
	//return

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
	//Create nodes
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

	n, k, f := 5, 3, 2
	m := GeneralizedWheelGenerator{}
	gu, err := m.Generate(n, k)
	if err != nil {
		fmt.Printf("invalid graph parameters: %v\n", err)
		os.Exit(1)
	}

	g := Directed(gu)
	//s, t := a, d

	start := rand.Intn(n)
	end := rand.Intn(n)
	for start == end {
		end = rand.Intn(n)
	}

	s, t := g.Node(int64(start)), g.Node(int64(end))

	// Print normal graph
	fmt.Printf("Graph (%v -> %v, via %v paths):\n", start, end, f)
	PrintGraphviz(g)
	/*
		digraph {
		    b -> a[label="0",weight="0",color=black,penwidth=1];
		    b -> d[label="1",weight="1",color=black,penwidth=1];
		    b -> c[label="0",weight="0",color=black,penwidth=1];
		    c -> a[label="1",weight="1",color=black,penwidth=1];
		    c -> d[label="0",weight="0",color=black,penwidth=1];
		    c -> b[label="0",weight="0",color=black,penwidth=1];
		    d -> b[label="1",weight="1",color=black,penwidth=1];
		    d -> c[label="0",weight="0",color=black,penwidth=1];
		    a -> b[label="0",weight="0",color=black,penwidth=1];
		    a -> c[label="1",weight="1",color=black,penwidth=1];
		}
	*/

	// Get shortest path
	//path, _ := ShortestPath(g, s, t)

	// Show first shortest path
	//fmt.Println("Naive path:")
	//PrintGraphvizHighlightPaths(g, []Path{path})
	/*
		digraph {
		    b -> a[label="0",weight="0",color=black,penwidth=1];
		    b -> d[label="1",weight="1",color=black,penwidth=1];
		    b -> c[label="0",weight="0",color=darkseagreen,penwidth=3];
		    c -> a[label="1",weight="1",color=black,penwidth=1];
		    c -> d[label="0",weight="0",color=darkseagreen,penwidth=3];
		    c -> b[label="0",weight="0",color=black,penwidth=1];
		    d -> b[label="1",weight="1",color=black,penwidth=1];
		    d -> c[label="0",weight="0",color=black,penwidth=1];
		    a -> b[label="0",weight="0",color=darkseagreen,penwidth=3];
		    a -> c[label="1",weight="1",color=black,penwidth=1];
		}
	*/

	edges, err := DisjointEdges(g, nil, s, t, f, nil, false)
	if err != nil {
		fmt.Printf("unable to find disjoint edges: %v\n", err)
		os.Exit(1)
	}

	//fmt.Println("All edges:")
	//PrintGraphvizHighlightPaths(g, []Path{edges})
	/*
		digraph {
		    b -> a[label="0",weight="0",color=black,penwidth=1];
		    b -> d[label="1",weight="1",color=darkseagreen,penwidth=3];
		    b -> c[label="0",weight="0",color=darkseagreen,penwidth=3];
		    c -> a[label="1",weight="1",color=black,penwidth=1];
		    c -> d[label="0",weight="0",color=darkseagreen,penwidth=3];
		    c -> b[label="0",weight="0",color=darkseagreen,penwidth=3];
		    d -> b[label="1",weight="1",color=black,penwidth=1];
		    d -> c[label="0",weight="0",color=black,penwidth=1];
		    a -> b[label="0",weight="0",color=darkseagreen,penwidth=3];
		    a -> c[label="1",weight="1",color=darkseagreen,penwidth=3];
		}
	*/

	filtered := FilterCounterparts(edges)
	//fmt.Println("Filtered edges:")
	//PrintGraphvizHighlightPaths(g, []Path{filtered})
	/*
		digraph {
		    b -> a[label="0",weight="0",color=black,penwidth=1];
		    b -> d[label="1",weight="1",color=mediumspringgreen,penwidth=3];
		    b -> c[label="0",weight="0",color=black,penwidth=1];
		    c -> a[label="1",weight="1",color=black,penwidth=1];
		    c -> d[label="0",weight="0",color=mediumspringgreen,penwidth=3];
		    c -> b[label="0",weight="0",color=black,penwidth=1];
		    d -> b[label="1",weight="1",color=black,penwidth=1];
		    d -> c[label="0",weight="0",color=black,penwidth=1];
		    a -> b[label="0",weight="0",color=mediumspringgreen,penwidth=3];
		    a -> c[label="1",weight="1",color=mediumspringgreen,penwidth=3];
		}
	*/

	paths := BuildPaths(filtered, s, t, k)
	fmt.Printf("Result (%v -> %v, via %v paths, valid: %v):\n", start, end, f, VerifySolution(g, s, t, f, paths))
	PrintGraphvizHighlightPaths(g, paths)
	/*
		digraph {
		    b -> c[label="0",weight="0",color=black,penwidth=1];
		    b -> a[label="0",weight="0",color=black,penwidth=1];
		    b -> d[label="1",weight="1",color=mediumspringgreen,penwidth=3];
		    c -> b[label="0",weight="0",color=black,penwidth=1];
		    c -> a[label="1",weight="1",color=black,penwidth=1];
		    c -> d[label="0",weight="0",color=violet,penwidth=3];
		    d -> b[label="1",weight="1",color=black,penwidth=1];
		    d -> c[label="0",weight="0",color=black,penwidth=1];
		    a -> b[label="0",weight="0",color=mediumspringgreen,penwidth=3];
		    a -> c[label="1",weight="1",color=violet,penwidth=3];
		}
	*/

	// Apply full first round of algo
	//g2 := NodeSplitting(g, s, t)
	//path2, _ := ShortestPath(g2, s, t)
	//fmt.Println("Round 1 algo:")
	//PrintGraphvizHighlightPaths(g2, []Path{path2})
	/*
		digraph {
		    b_out -> c_in[label="0",weight="0",color=slateblue,penwidth=3];
		    b_out -> d[label="1",weight="1",color=black,penwidth=1];
		    b_out -> a[label="0",weight="0",color=black,penwidth=1];
		    c_in -> c_out[label="0",weight="0",color=slateblue,penwidth=3];
		    c_out -> a[label="1",weight="1",color=black,penwidth=1];
		    c_out -> d[label="0",weight="0",color=slateblue,penwidth=3];
		    c_out -> b_in[label="0",weight="0",color=black,penwidth=1];
		    a -> b_in[label="0",weight="0",color=slateblue,penwidth=3];
		    a -> c_in[label="1",weight="1",color=black,penwidth=1];
		    b_in -> b_out[label="0",weight="0",color=slateblue,penwidth=3];
		    d -> b_in[label="1",weight="1",color=black,penwidth=1];
		    d -> c_in[label="0",weight="0",color=black,penwidth=1];
		}
	*/

	//for _, e := range path2 {
	//	InverseLink(g2, e)
	//}

	// Apply full second round of algo
	//path3, _ := ShortestPath(g2, s, t)
	//fmt.Println("Round 2 algo:")
	//PrintGraphvizHighlightPaths(g2, []Path{path3})
	/*
		digraph {
		    b_out -> b_in[label="-0",weight="-0",color=black,penwidth=1];
		    b_out -> d[label="1",weight="1",color=slateblue,penwidth=3];
		    b_out -> a[label="0",weight="0",color=black,penwidth=1];
		    c_in -> b_out[label="-0",weight="-0",color=slateblue,penwidth=3];
		    c_out -> a[label="1",weight="1",color=black,penwidth=1];
		    c_out -> c_in[label="-0",weight="-0",color=black,penwidth=1];
		    c_out -> b_in[label="0",weight="0",color=black,penwidth=1];
		    a -> c_in[label="1",weight="1",color=slateblue,penwidth=3];
		    b_in -> a[label="-0",weight="-0",color=black,penwidth=1];
		    d -> c_in[label="0",weight="0",color=black,penwidth=1];
		    d -> b_in[label="1",weight="1",color=black,penwidth=1];
		    d -> c_out[label="-0",weight="-0",color=black,penwidth=1];
		}
	*/

	lookup, err := BuildLookupTable(gu, s, k, 0, false)
	if err != nil {
		fmt.Printf("failed to build lookup table: %v\n", err)
		os.Exit(1)
	}

	for to, paths := range lookup {
		fmt.Printf("%v -> %v\n", to, paths)
	}

	fmt.Printf("valid %v-disjoint paths: %v", f, VerifyDisjointPaths(paths, s, t, f))
}

func alt() {
	gr := simple.NewWeightedUndirectedGraph(0, 0)

	a := NewNodeUndirected(gr, "a")
	gr.AddNode(a)

	b := NewNodeUndirected(gr, "b")
	gr.AddNode(b)

	c := NewNodeUndirected(gr, "c")
	gr.AddNode(c)

	d := NewNodeUndirected(gr, "d")
	gr.AddNode(d)

	e := NewNodeUndirected(gr, "e")
	gr.AddNode(e)

	f := NewNodeUndirected(gr, "f")
	gr.AddNode(f)

	g := NewNodeUndirected(gr, "g")
	gr.AddNode(g)

	h := NewNodeUndirected(gr, "h")
	gr.AddNode(h)

	// Create edges
	ab := gr.NewWeightedEdge(a, b, 1)
	gr.SetWeightedEdge(ab)

	bc := gr.NewWeightedEdge(b, c, 1)
	gr.SetWeightedEdge(bc)

	cd := gr.NewWeightedEdge(c, d, 1)
	gr.SetWeightedEdge(cd)

	dh := gr.NewWeightedEdge(d, h, 1)
	gr.SetWeightedEdge(dh)

	ae := gr.NewWeightedEdge(a, e, 1)
	gr.SetWeightedEdge(ae)

	ed := gr.NewWeightedEdge(e, d, 1)
	gr.SetWeightedEdge(ed)

	ef := gr.NewWeightedEdge(e, f, 3)
	gr.SetWeightedEdge(ef)

	fd := gr.NewWeightedEdge(f, d, 1)
	gr.SetWeightedEdge(fd)

	fg := gr.NewWeightedEdge(f, g, 1)
	gr.SetWeightedEdge(fg)

	gh := gr.NewWeightedEdge(g, h, 1)
	gr.SetWeightedEdge(gh)

	ah := gr.NewWeightedEdge(a, h, 6)
	gr.SetWeightedEdge(ah)

	gd := Directed(gr)

	// Print normal graph
	fmt.Println("Graph:")
	PrintGraphviz(gd)

	k := 3
	s, t := a, h
	edges, err := DisjointEdges(gd, nil, s, t, k, nil, false)
	if err != nil {
		fmt.Printf("unable to find disjoint edges: %v\n", err)
		os.Exit(1)
	}

	filtered := FilterCounterparts(edges)
	fmt.Println("Filtered edges:")
	PrintGraphvizHighlightPaths(gd, []Path{filtered})

	paths := BuildPaths(filtered, s, t, k)
	fmt.Println("result:")
	PrintGraphvizHighlightPaths(gd, paths)
}

func BuildLookupTable(gu *simple.WeightedUndirectedGraph, s graph.Node, k int, w int, skipNeighbour bool) (map[uint64][]Path, error) {
	res := make(map[uint64][]Path)
	g := Directed(gu)

	nodes := gu.Nodes()
	orderedNodes := make([]graph.Node, 0, nodes.Len())

	split := NodeSplitting(g)
	additionalWeight := make([][]int, len(split.nodes)+1)

	for nodes.Next() {
		n := nodes.Node()
		orderedNodes = append(orderedNodes, n)
	}

	// Set initial weights to w
	for i := 0; i < len(split.nodes); i++ {
		additionalWeight[i] = make([]int, len(split.nodes)+1)
	}

	edg := split.g.Edges()
	for edg.Next() {
		e := edg.Edge().(simple.WeightedEdge)
		if e.W < 0.5 {
			continue
		}

		additionalWeight[uint64(e.From().ID())][uint64(e.To().ID())] = w
	}

	sort.Slice(orderedNodes, func(i, j int) bool {
		return orderedNodes[i].ID() < orderedNodes[j].ID()
	})

	for _, n := range orderedNodes {
		// No lookup to self needed
		if n.ID() == s.ID() {
			continue
		}

		paths, err := DisjointPaths(g, split, s, n, k, additionalWeight, skipNeighbour)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build paths to %v", n)
		}

		for _, p := range paths {
			for _, e := range p {
				// Path is now used, update additional weight to zero
				additionalWeight[uint64(e.From().ID())][uint64(e.To().ID())] = 0
			}
		}

		res[uint64(n.ID())] = paths
		//fmt.Printf("done: %v\n", n.ID())
	}

	return res, nil
}

func DisjointPaths(g *simple.WeightedDirectedGraph, split *SplitGraph, s, t graph.Node, k int, additionalWeight [][]int, neighbourHop bool) ([]Path, error) {
	res, err := DisjointEdges(g, split, s, t, k, additionalWeight, neighbourHop)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find disjoint edges")
	}

	filtered := FilterCounterparts(res)

	return BuildPaths(filtered, s, t, k), nil
}

func DisjointEdges(g *simple.WeightedDirectedGraph, split *SplitGraph, s, t graph.Node, k int, additionalWeight [][]int, neighbourHop bool) ([]graph.WeightedEdge, error) {
	// If direct neighbour hopping is used, check if that can be used
	if neighbourHop && g.HasEdgeFromTo(s.ID(), t.ID()) {
		return []graph.WeightedEdge{g.WeightedEdge(s.ID(), t.ID())}, nil
	}

	if split == nil {
		split = NodeSplitting(g)
	}
	res := make([]graph.WeightedEdge, 0, k)

	source, sink := s.ID()+int64(len(split.nodes)/2), t.ID()
	edges := FindAdjMap(split.g, int64(len(split.nodes)))

	for i := 0; i < k; i++ {
		path, err := ShortestPath(source, sink, edges, additionalWeight)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to find %vnd path", k)
		}

		//fmt.Printf("round %v of algo:\n", i)
		//PrintGraphvizHighlightPaths(g2, [][]graph.WeightedEdge{path})

		// Inverse links
		for _, e := range path {
			// Add edge to path if this is not an internal transfer (in->out)
			if e.To().(Node).original.ID() != e.From().(Node).original.ID() {
				res = append(res, split.g.NewWeightedEdge(e.From().(Node).original, e.To().(Node).original, e.Weight()))
			}

			// Update adj map
			edges[e.To().ID()][e.From().ID()] = g.NewWeightedEdge(e.To(), e.From(), e.Weight()*-1)
			edges[e.From().ID()][e.To().ID()] = nil
		}
	}

	return res, nil
}

func FindAdjMap(g *simple.WeightedDirectedGraph, max int64) [][]graph.WeightedEdge {
	edges := g.Edges()
	res := make([][]graph.WeightedEdge, max+1)

	for edges.Next() {
		e := edges.Edge().(graph.WeightedEdge)

		if res[e.From().ID()] == nil {
			res[e.From().ID()] = make([]graph.WeightedEdge, max+1)
		}

		res[e.From().ID()][e.To().ID()] = e
	}

	return res
}

func FilterCounterparts(edges []graph.WeightedEdge) []graph.WeightedEdge {
	res := make([]graph.WeightedEdge, 0, len(edges))
	drop := make(map[int64][]int64)

	// First do a drop run
	for _, e := range edges {
		if d, ok := drop[e.To().ID()]; ok {
			drop[e.To().ID()] = append(d, e.From().ID())
		} else {
			drop[e.To().ID()] = []int64{e.From().ID()}
		}
	}

	// Then keep the edges which have no counterpart
next:
	for _, e := range edges {
		drops := drop[e.From().ID()]
		t := e.To().ID()

		for _, d := range drops {
			if d == t {
				continue next
			}
		}

		res = append(res, e)
	}

	return res
}

func BuildPaths(edges []graph.WeightedEdge, s, t graph.Node, k int) []Path {
	res := make([]Path, 0, k)
	next := make(map[int64]graph.WeightedEdge)

	for _, e := range edges {
		// Exclude sink
		if e.From().ID() == t.ID() {
			continue
		}

		next[e.From().ID()] = e
	}

	for _, e := range edges {
		// Early stop, if enough paths have been built
		if len(res) == k {
			break
		}

		// Only do this for starting edges
		if e.From().ID() == s.ID() {
			path := make([]graph.WeightedEdge, 0)

			cur := e

			for cur.To().ID() != t.ID() {
				path = append(path, cur)

				cur = next[cur.To().ID()]
			}

			path = append(path, cur)
			res = append(res, path)
		}
	}

	return res
}

type SplitGraph struct {
	g     *simple.WeightedDirectedGraph
	nodes []int64
}

func NodeSplitting(g *simple.WeightedDirectedGraph) *SplitGraph {
	nodes := g.Nodes()
	cnt := nodes.Len()

	g2 := simple.NewWeightedDirectedGraph(0, 0)
	res := make([]int64, 0, cnt*2)

	outMap := make(map[string]graph.Node)
	inMap := make(map[string]graph.Node)

	// Add splitted nodes
	for nodes.Next() {
		n := nodes.Node()

		// Add new nodes and copies
		name := n.(Node).Name

		inNode := NewNodeSplit(name+"_in", n, true, int64(cnt))
		g2.AddNode(inNode)
		inMap[name] = inNode
		outNode := NewNodeSplit(name+"_out", n, false, int64(cnt))
		g2.AddNode(outNode)
		outMap[name] = outNode

		// Get all edges
		g2.SetWeightedEdge(g2.NewWeightedEdge(inNode, outNode, 0))
		res = append(res, inNode.ID(), outNode.ID())
	}

	nodes.Reset()

	// Re-add edges
	for nodes.Next() {
		n := nodes.Node()
		name := n.(Node).Name
		in, out := g.To(n.ID()), g.From(n.ID())

		for in.Next() {
			f := in.Node()
			w := g.WeightedEdge(f.ID(), n.ID()).Weight()

			g2.SetWeightedEdge(g2.NewWeightedEdge(outMap[f.(Node).Name], inMap[name], w))
		}

		for out.Next() {
			t := out.Node()
			w := g.WeightedEdge(n.ID(), t.ID()).Weight()

			g2.SetWeightedEdge(g2.NewWeightedEdge(outMap[name], inMap[t.(Node).Name], w))
		}
	}

	return &SplitGraph{
		g:     g2,
		nodes: res,
	}
}
