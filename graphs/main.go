package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"math/rand"
	"os"
	"time"
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

func NewNode(g *simple.WeightedDirectedGraph, name string) graph.Node {
	return Node{
		id:   g.NewNode().ID(),
		name: name,
	}
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

func main() {
	//alt()
	//return

	gr := simple.NewWeightedUndirectedGraph(0, 0)

	/*        b
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

	g := Directed(gr)

	// Print normal graph
	fmt.Println("Graph:")
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
	path, _ := ShortestPath(g, a, d)

	// Show first shortest path
	fmt.Println("Naive path:")
	PrintGraphvizHighlightPaths(g, []Path{path})
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

	k := 2
	edges, err := DisjointEdges(g, a, d, k)
	if err != nil {
		fmt.Printf("unable to find disjoint edges: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All edges:")
	PrintGraphvizHighlightPaths(g, []Path{edges})
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
	fmt.Println("Filtered edges:")
	PrintGraphvizHighlightPaths(g, []Path{filtered})
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

	paths := BuildPaths(filtered, a, d, k)
	fmt.Println("result:")
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
	g2 := NodeSplitting(g, a, d)
	path2, _ := ShortestPath(g2, a, d)
	fmt.Println("Round 1 algo:")
	PrintGraphvizHighlightPaths(g2, []Path{path2})
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

	for _, e := range path2 {
		InverseLink(g2, e)
	}

	// Apply full second round of algo
	path3, _ := ShortestPath(g2, a, d)
	fmt.Println("Round 2 algo:")
	PrintGraphvizHighlightPaths(g2, []Path{path3})
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
	edges, err := DisjointEdges(gd, s, t, k)
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

func DisjointEdges(g *simple.WeightedDirectedGraph, s, t graph.Node, k int) ([]graph.WeightedEdge, error) {
	g2 := NodeSplitting(g, s, t)
	res := make([]graph.WeightedEdge, 0, k)

	for i := 0; i < k; i++ {
		path, err := ShortestPath(g2, s, t)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to find %vnd path", k)
		}

		//fmt.Printf("round %v of algo:\n", i)
		//PrintGraphvizHighlightPaths(g2, [][]graph.WeightedEdge{path})

		// Inverse links
		for _, e := range path {
			// Add edge to path if this is not an internal transfer (in->out)
			if e.To().(Node).original.ID() != e.From().(Node).original.ID() {
				res = append(res, g2.NewWeightedEdge(e.From().(Node).original, e.To().(Node).original, e.Weight()))
			}

			InverseLink(g2, e)
		}
	}

	return res, nil
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

// Bellman-Ford (needs to be capable of handling negative weights)
func ShortestPath(g *simple.WeightedDirectedGraph, s, t graph.Node) (Path, error) {
	nodes := g.Nodes()

	distance := make(map[int64]float64)
	predecessor := make(map[int64]int64)

	// Step 1: Init graph
	for nodes.Next() {
		n := nodes.Node()
		distance[n.ID()] = math.MaxInt64
		predecessor[n.ID()] = -1
	}

	nodes.Reset()
	distance[s.ID()] = 0

	// Step 2: Relax edges repeatedly
	for i := 0; i < nodes.Len(); i++ {
		edges := g.Edges()

		for edges.Next() {
			e := edges.Edge().(graph.WeightedEdge)

			if d := distance[e.From().ID()] + e.Weight(); d < distance[e.To().ID()] {
				distance[e.To().ID()] = d
				predecessor[e.To().ID()] = e.From().ID()
			}
		}
	}

	// Step 3: Check for negative-weight cycles -> TODO: probably not useful/needed here

	// Step 4: build path to target
	res := make([]graph.WeightedEdge, 0, nodes.Len())
	cur := t.ID()

	for cur != s.ID() {
		next := predecessor[cur]

		if next == -1 {
			return nil, errors.New("no path")
		}

		res = append(res, g.WeightedEdge(next, cur))
		cur = next
	}

	return res, nil
}

func NodeSplitting(g *simple.WeightedDirectedGraph, s, t graph.Node) *simple.WeightedDirectedGraph {
	g2 := simple.NewWeightedDirectedGraph(0, 0)

	s, t = Node{
		id:       s.ID(),
		name:     s.(Node).name,
		original: s,
	}, Node{
		id:       t.ID(),
		name:     t.(Node).name,
		original: t,
	}

	g2.AddNode(s)
	g2.AddNode(t)
	outMap := make(map[string]graph.Node)
	inMap := make(map[string]graph.Node)

	inMap[s.(Node).name] = s
	outMap[s.(Node).name] = s

	inMap[t.(Node).name] = t
	outMap[t.(Node).name] = t

	nodes := g.Nodes()
	// Add splitted nodes
	for nodes.Next() {
		n := nodes.Node()

		// Source and sink do not have to be split
		if n.ID() == s.ID() || n.ID() == t.ID() {
			continue
		}

		// Add new nodes and copies
		name := n.(Node).name

		inNode := NewNodeSplit(g2, name+"_in", n)
		g2.AddNode(inNode)
		inMap[name] = inNode
		outNode := NewNodeSplit(g2, name+"_out", n)
		g2.AddNode(outNode)
		outMap[name] = outNode

		// Get all edges
		g2.SetWeightedEdge(g2.NewWeightedEdge(inNode, outNode, 0))
	}

	nodes.Reset()

	// Re-add edges
	for nodes.Next() {
		n := nodes.Node()
		name := n.(Node).name
		in, out := g.To(n.ID()), g.From(n.ID())

		for in.Next() {
			f := in.Node()
			g2.SetWeightedEdge(g2.NewWeightedEdge(outMap[f.(Node).name], inMap[name], g.WeightedEdge(f.ID(), n.ID()).Weight()))
		}

		for out.Next() {
			t := out.Node()
			g2.SetWeightedEdge(g2.NewWeightedEdge(outMap[name], inMap[t.(Node).name], g.WeightedEdge(n.ID(), t.ID()).Weight()))
		}
	}

	return g2
}

func InverseLink(g *simple.WeightedDirectedGraph, e graph.WeightedEdge) {
	w := e.Weight() * -1
	f, t := e.To(), e.From()
	fid, tid := f.ID(), t.ID()

	// Remove edges
	g.RemoveEdge(fid, tid)
	g.RemoveEdge(tid, fid)

	// Add inverse edge
	g.SetWeightedEdge(g.NewWeightedEdge(f, t, w))
}

// Helper to make life easier
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

// Pretty printing of graphs (pasted from graphviz wiki, slightly curated)
var colors = []string{
	"aquamarine",
	"bisque",
	"blue",
	"blueviolet",
	"brown",
	"burlywood",
	"cadetblue",
	"chartreuse",
	"chocolate",
	"coral",
	"cornflowerblue",
	"crimson",
	"cyan",
	"darkgoldenrod",
	"darkgreen",
	"darkkhaki",
	"darkolivegreen",
	"darkorange",
	"darkorchid",
	"darksalmon",
	"darkseagreen",
	"darkslateblue",
	"darkturquoise",
	"darkviolet",
	"deeppink",
	"deepskyblue",
	"dimgray",
	"dodgerblue",
	"firebrick",
	"firebrick",
	"forestgreen",
	"gold",
	"goldenrod",
	"grey",
	"green",
	"greenyellow",
	"hotpink",
	"indianred",
	"indigo",
	"khaki",
	"lawngreen",
	"lightblue",
	"lightcoral",
	"lightgray",
	"lightpink",
	"lightsalmon",
	"lightseagreen",
	"lightskyblue",
	"lightslategray",
	"lightsteelblue",
	"limegreen",
	"magenta",
	"maroon",
	"mediumaquamarine",
	"mediumblue",
	"mediumorchid",
	"mediumpurple",
	"mediumseagreen",
	"mediumslateblue",
	"mediumspringgreen",
	"mediumturquoise",
	"mediumvioletred",
	"midnightblue",
	"navajowhite",
	"navy",
	"olivedrab",
	"orange",
	"orangered",
	"orchid",
	"palegoldenrod",
	"palegreen",
	"paleturquoise",
	"palevioletred",
	"peachpuff",
	"peru",
	"pink",
	"plum",
	"powderblue",
	"purple",
	"red",
	"rosybrown",
	"royalblue",
	"saddlebrown",
	"salmon",
	"sandybrown",
	"seagreen",
	"sienna",
	"skyblue",
	"slateblue",
	"slategray",
	"springgreen",
	"steelblue",
	"tan",
	"thistle",
	"tomato",
	"turquoise",
	"violet",
	"wheat",
	"yellow",
	"yellowgreen",
}

func PrintGraphviz(g graph.WeightedDirected) {
	PrintGraphvizHighlightPaths(g, []Path{})
}

func PrintGraphvizHighlightPaths(g graph.WeightedDirected, paths []Path) {
	nodes := g.Nodes()
	fmt.Printf("digraph {\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	cols := make(map[int64]map[int64]string)

	for _, p := range paths {
		col := colors[r.Intn(len(colors))]

		for _, e := range p {
			if _, ok := cols[e.From().ID()]; !ok {
				cols[e.From().ID()] = make(map[int64]string)
			}

			cols[e.From().ID()][e.To().ID()] = col
		}
	}

	for nodes.Next() {
		n := nodes.Node()
		col, ok := cols[n.ID()]

		to := g.From(n.ID())
		for to.Next() {
			t := to.Node()
			w := g.WeightedEdge(n.ID(), t.ID()).Weight()

			c := "black"
			width := 1
			if ok {
				if color, ok := col[t.ID()]; ok {
					c = color
					width = 3
				}
			}

			fmt.Printf("    %v -> %v[label=\"%v\",weight=\"%v\",color=%v,penwidth=%v];\n", n, t, w, w, c, width)
		}
	}

	fmt.Printf("}\n")
}
