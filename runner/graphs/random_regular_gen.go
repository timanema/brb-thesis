package graphs

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"math/rand"
	"runtime"
	"time"
)

// https://mediatum.ub.tum.de/doc/1315533/file.pdf
// https://github.com/networkx/networkx/blob/fec20e0a6767eca1b40f19e1cf06387cdaea0f13/networkx/generators/random_graphs.py#L491
type RandomRegularGenerator struct{}

func (r RandomRegularGenerator) checkConnected(g *simple.WeightedUndirectedGraph, k int) bool {
	return FindConnectedness(g) >= k
}

func suitable(edges map[edge]struct{}, potentialEdges map[int64]int64) bool {
	if len(potentialEdges) == 0 {
		return true
	}

	// Get keys, so order is identical for both loops
	potKeys := make([]int64, 0, len(potentialEdges))
	for k := range potentialEdges {
		potKeys = append(potKeys, k)
	}

	for _, s1 := range potKeys {
		for _, s2 := range potKeys {
			if s1 == s2 {
				break
			}

			if s1 > s2 {
				s1, s2 = s2, s1
			}

			if _, ok := edges[edge{T: s1, F: s2}]; !ok {
				return true
			}
		}
	}

	return false
}

type edge struct {
	T, F int64
}

func makePairs(xs []int64) []pair {
	res := make([]pair, 0, len(xs)/2)

	for i := 0; i < len(xs); i++ {
		a := xs[i]
		i += 1

		res = append(res, pair{a: a, b: xs[i]})
	}

	return res
}

func tryCreation(ctx context.Context, n, d int) map[edge]struct{} {
	rand.Seed(time.Now().Unix())

	edges := make(map[edge]struct{})
	stubs := make([]int64, 0, n*d)
	for i := 0; i < d; i++ {
		l := make([]int64, n)
		for j := int64(0); j < int64(n); j++ {
			l[j] = j
		}

		stubs = append(stubs, l...)
	}

	for len(stubs) > 0 {
		if ctx.Err() != nil {
			return nil
		}

		potentialEdges := make(map[int64]int64)
		rand.Shuffle(len(stubs), func(i, j int) {
			stubs[i], stubs[j] = stubs[j], stubs[i]
		})

		pairs := makePairs(stubs)
		for _, p := range pairs {
			if p.a > p.b {
				p.a, p.b = p.b, p.a
			}

			if _, ok := edges[edge{T: p.a, F: p.b}]; p.a != p.b && !ok {
				edges[edge{T: p.a, F: p.b}] = struct{}{}
			} else {
				potentialEdges[p.a] += 1
				potentialEdges[p.b] += 1
			}
		}

		if !suitable(edges, potentialEdges) {
			return nil
		}

		stubs = nil
		for node, potential := range potentialEdges {
			for i := int64(0); i < potential; i++ {
				stubs = append(stubs, node)
			}
		}
	}

	return edges
}

func (r RandomRegularGenerator) Generate(n, k, d int) (*simple.WeightedUndirectedGraph, error) {
	if n*d%2 != 0 {
		return nil, errors.Errorf("n*d must be even: n=%v, d=%v", n, d)
	}

	if d < k {
		return nil, errors.Errorf("degree cannot be lower than connectivity: k=%v, d=%v", k, d)
	}

	runners := runtime.NumCPU()
	ctx, cancel := context.WithCancel(context.TODO())
	res := make(chan *simple.WeightedUndirectedGraph, runners)

	for i := 0; i < runners; i++ {
		go func() {
			for {
				if ctx.Err() != nil {
					return
				}

				edges := tryCreation(ctx, n, d)
				if edges == nil {
					continue
				}

				g := simple.NewWeightedUndirectedGraph(0, 0)
				for e := range edges {
					g.SetWeightedEdge(g.NewWeightedEdge(node(g, int(e.F)), node(g, int(e.T)), 1))
				}

				if c := FindConnectedness(g); c >= k {
					res <- g
				} else {
					fmt.Printf("found random graph, connectivity too low: %v < %v\n", c, k)
				}
			}
		}()
	}

	g := <-res
	cancel()
	return g, nil
}

func (r RandomRegularGenerator) Cache() (bool, string) {
	return true, "random"
}
