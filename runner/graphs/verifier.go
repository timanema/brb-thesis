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

func bfs(edges [][]graph.WeightedEdge, s, t int64) Path {
	visited := make([]bool, len(edges))
	parent := make([]int64, len(edges))
	queue := make([]int64, 0)

	queue = append(queue, s)
	visited[s] = true

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		for _, e := range edges[n] {
			if e == nil {
				continue
			}
			adj := e.To().ID()

			if !visited[adj] && e.Weight() > 0 {
				queue = append(queue, adj)
				visited[adj] = true
				parent[adj] = n + 1
			}
		}
	}

	res := make([]graph.WeightedEdge, 0)
	cur := t

	for cur != s {
		next := parent[cur]
		if next == 0 {
			return nil
		}

		res = append(res, edges[next-1][cur])
		cur = next - 1
	}

	return res
}

func maxFlow(edges [][]graph.WeightedEdge, s, t int64) int {
	if s == t {
		fmt.Println("Should not happen: s == t for maxFlow()")
		return -1
	}

	flow := 0
	for {
		p := bfs(edges, s, t)
		if p == nil || len(p) == 0 {
			return flow
		}
		flow += 1

		for _, e := range p {
			edges[e.From().ID()][e.To().ID()] = simple.WeightedEdge{F: e.From(), T: e.To(), W: e.Weight() - 1}
			edges[e.To().ID()][e.From().ID()] = simple.WeightedEdge{F: e.To(), T: e.From(), W: e.Weight() + 1}
		}
	}
}

// Verifies there are k disjoint paths through the graph
func VerifyDisjointPaths(paths []Path, s, t graph.Node, k int) bool {
	// TODO: reuse aux data structures
	g := simple.NewWeightedDirectedGraph(0, 0)
	max := s.ID()

	if id := t.ID(); id > max {
		max = id
	}

	for _, path := range paths {
		for _, e := range path {
			g.SetWeightedEdge(g.NewWeightedEdge(e.From(), e.To(), 1))
			g.SetWeightedEdge(g.NewWeightedEdge(e.To(), e.From(), 1))

			if id := e.To().ID(); id > max {
				max = id
			}

			if id := e.From().ID(); id > max {
				max = id
			}
		}
	}

	edges := FindAdjMap(g, max)

	return maxFlow(edges, s.ID(), t.ID()) >= k
}
