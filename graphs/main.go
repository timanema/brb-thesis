package main

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"math"
	"math/rand"
	"time"
)

type Node struct {
	id   int64
	name string
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

func main() {
	g := simple.NewWeightedDirectedGraph(0, 0)

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
	a := NewNode(g, "a")
	g.AddNode(a)

	b := NewNode(g, "b")
	g.AddNode(b)

	c := NewNode(g, "c")
	g.AddNode(c)

	d := NewNode(g, "d")
	g.AddNode(d)

	// Create edges
	ab := g.NewWeightedEdge(a, b, 0)
	abR := g.NewWeightedEdge(b, a, 0)
	g.SetWeightedEdge(ab)
	g.SetWeightedEdge(abR)

	ac := g.NewWeightedEdge(a, c, 1)
	acR := g.NewWeightedEdge(c, a, 1)
	g.SetWeightedEdge(ac)
	g.SetWeightedEdge(acR)

	bd := g.NewWeightedEdge(b, d, 1)
	bdR := g.NewWeightedEdge(d, b, 1)
	g.SetWeightedEdge(bd)
	g.SetWeightedEdge(bdR)

	cd := g.NewWeightedEdge(c, d, 0)
	cdR := g.NewWeightedEdge(d, c, 0)
	g.SetWeightedEdge(cd)
	g.SetWeightedEdge(cdR)

	bc := g.NewWeightedEdge(b, c, 0)
	bcR := g.NewWeightedEdge(c, b, 0)
	g.SetWeightedEdge(bc)
	g.SetWeightedEdge(bcR)

	// Print normal graph
	PrintGraphviz(g)

	// Get shortest path
	path := ShortestPath(g, a, d)
	fmt.Println(path)

	// Show first shortest path
	PrintGraphvizHighlightPath(g, path)

	// Apply full first round of algo
	g2 := NodeSplitting(g, a, d)
	path2 := ShortestPath(g2, a, d)
	fmt.Println("round 1 of algo:")
	PrintGraphvizHighlightPath(g2, path2)

	for _, e := range path2 {
		InverseLink(g2, e)
	}
	PrintGraphviz(g2)
}

// Bellman-Ford (needs to be capable of handling negative weights)
func ShortestPath(g *simple.WeightedDirectedGraph, s, t graph.Node) []graph.WeightedEdge {
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

		// TODO: check for -1 -> no path

		res = append(res, g.WeightedEdge(next, cur))
		cur = next
	}

	return res
}

func NodeSplitting(g *simple.WeightedDirectedGraph, s, t graph.Node) *simple.WeightedDirectedGraph {
	g2 := simple.NewWeightedDirectedGraph(0, 0)
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

		inNode := NewNode(g2, name + "_in")
		g2.AddNode(inNode)
		inMap[name] = inNode
		outNode := NewNode(g2, name + "_out")
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
	"gainsboro",
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
	"lemonchiffon",
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
	PrintGraphvizHighlightPath(g, []graph.WeightedEdge{})
}

func PrintGraphvizHighlightPath(g graph.WeightedDirected, edges []graph.WeightedEdge) {
	nodes := g.Nodes()
	fmt.Printf("digraph {\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	col := colors[r.Intn(len(colors))]

	lookup := make(map[int64]int64)
	for _, e := range edges {
		lookup[e.From().ID()] = e.To().ID()
	}

	for nodes.Next() {
		n := nodes.Node()

		to := g.From(n.ID())
		for to.Next() {
			t := to.Node()
			w := g.WeightedEdge(n.ID(), t.ID()).Weight()

			if path, ok := lookup[n.ID()]; ok && path == t.ID() {
				fmt.Printf("    %v -> %v[label=\"%v\",weight=\"%v\",color=%v,penwidth=3.0];\n", n, t, w, w, col)
			} else {
				fmt.Printf("    %v -> %v[label=\"%v\",weight=\"%v\"];\n", n, t, w, w)
			}
		}
	}

	fmt.Printf("}\n")
}
