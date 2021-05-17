package graphs

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func VerifySolution(g graph.WeightedDirected, s, t graph.Node, k int, paths []Path) bool {
	// Check if enough paths
	if len(paths) < k {
		return false
	}

	// Verify the paths are disjoint
	seen := make(map[int64]struct{})

	for _, p := range paths {
		for _, e := range p {
			// Skip start node
			if e.From().ID() == s.ID() {
				continue
			}

			// Not disjoint
			if _, ok := seen[e.From().ID()]; ok {
				return false
			}

			seen[e.From().ID()] = struct{}{}
		}
	}

	// Verify the paths are valid
	for _, p := range paths {
		// Invalid path
		if len(p) < 1 {
			return false
		}

		prev := p[0]

		// All paths should start at s
		if prev.From().ID() != s.ID() {
			return false
		}

		for i, cur := range p {
			// Every edge should be valid
			if !g.HasEdgeBetween(cur.From().ID(), cur.To().ID()) {
				return false
			}

			// Previous .To() should equal current .From() (if this is not the first edge)
			if prev.To().ID() != cur.From().ID() && i != 0 {
				return false
			}

			prev = cur
		}

		// All paths should end at t
		if prev.To().ID() != t.ID() {
			return false
		}
	}

	return true
}

func bfs(g graph.WeightedDirected, s, t graph.Node) Path {
	visited := make(map[int64]struct{}, g.Nodes().Len())
	parent := make(map[int64]int64, g.Nodes().Len())
	queue := make([]int64, 0)

	queue = append(queue, s.ID())
	visited[s.ID()] = struct{}{}

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		adj := g.From(n)
		for adj.Next() {
			a := adj.Node().ID()
			w := g.WeightedEdge(n, a).Weight()

			if _, ok := visited[a]; !ok && w > 0 {
				queue = append(queue, a)
				visited[a] = struct{}{}
				parent[a] = n
			}
		}
	}

	res := make([]graph.WeightedEdge, 0)
	cur := t.ID()

	for cur != s.ID() {
		next, ok := parent[cur]
		if !ok {
			return nil
		}

		res = append(res, g.WeightedEdge(next, cur))
		cur = next
	}

	return res
}

func findDisjointPaths(g *simple.WeightedDirectedGraph, s, t graph.Node) int {
	if s.ID() == t.ID() {
		fmt.Println("Should not happen: s == t for findDisjointPaths()")
		return -1
	}

	flow := 0
	for {
		p := bfs(g, s, t)
		if p == nil || len(p) == 0 {
			return flow
		}
		flow += 1

		for _, e := range p {
			g.SetWeightedEdge(g.NewWeightedEdge(e.From(), e.To(), e.Weight()-1))
			g.SetWeightedEdge(g.NewWeightedEdge(e.To(), e.From(), e.Weight()+1))
		}
	}
}

// Verifies there are k disjoint paths through the graph
func VerifyDisjointPaths(paths []Path, s, t graph.Node, k int) bool {
	g := simple.NewWeightedDirectedGraph(0, 0)

	for _, path := range paths {
		for _, e := range path {
			g.SetWeightedEdge(g.NewWeightedEdge(e.From(), e.To(), 1))
			g.SetWeightedEdge(g.NewWeightedEdge(e.To(), e.From(), 1))
		}
	}

	return findDisjointPaths(g, s, t) >= k
}

type pair struct {
	a, b int64
}

func findPairs(xs []int64) []pair {
	res := make([]pair, 0, len(xs)*len(xs))

	for i := 0; i < len(xs); i++ {
		for j := 0; j < len(xs); j++ {
			if i == j {
				continue
			}

			res = append(res, pair{a: xs[i], b: xs[j]})
		}
	}

	return res
}

func FindConnectedness(g graph.Undirected) int {
	// TODO: implement
	return 0
}
