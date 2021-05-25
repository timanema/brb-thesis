package algo

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/graphs"
)

type Path struct {
	P    graphs.Path
	Prio bool
}
type RoutingTable map[uint64][]Path
type BroadcastPlan map[uint64][]Path

type DolevPath struct {
	Desired, Actual graphs.Path
	Prio            bool
}

func combinePaths(paths []Path) BroadcastPlan {
	res := make(BroadcastPlan)

	for _, p := range paths {
		next := uint64(p.P[0].To().ID())
		res[next] = append(res[next], p)
	}

	return res
}

func BuildRoutingTable(g *simple.WeightedUndirectedGraph, s graph.Node, k int, w int, skipNeighbour bool) (RoutingTable, error) {
	routes, err := graphs.BuildLookupTable(g, s, k, w, skipNeighbour)
	if err != nil {
		return nil, err
	}

	rt := make(RoutingTable, len(routes))

	for dst, paths := range routes {
		res := make([]Path, 0, len(paths))

		// TODO: remove automatic priority after deadlock detection works
		for _, p := range paths {
			res = append(res, Path{
				P: p,
				//Prio: true,
			})
		}

		rt[dst] = res
	}

	return rt, nil
}

func FilterSubpaths(r RoutingTable) {
	for dst, paths := range r {
		res := make([]Path, 0, len(paths))

		for i, p := range paths {
			sub := false

		subSearch:
			for otherDst, otherPaths := range r {
				for j, p2 := range otherPaths {
					if dst == otherDst && i == j {
						continue
					}

					if graphs.IsSubPath(p.P, p2.P) {
						sub = true
						break subSearch
					}
				}
			}

			if !sub {
				res = append(res, p)
			}
		}

		r[dst] = res
	}
}

func DolevRouting(r RoutingTable, combine, filter bool) BroadcastPlan {
	if filter {
		FilterSubpaths(r)
	}

	br := make([]Path, 0, len(r))
	for _, g := range r {
		br = append(br, g...)
	}

	if combine {
		return combinePaths(br)
	}

	res := make(BroadcastPlan)
	for _, p := range br {
		res[uint64(p.P[0].To().ID())] = append(res[uint64(p.P[0].To().ID())], p)
	}

	return res
}

func CombineDolevPaths(paths []DolevPath) map[uint64][]DolevPath {
	res := make(map[uint64][]DolevPath)

	for _, p := range paths {
		if cur := len(p.Actual); len(p.Desired) > cur {
			next := uint64(p.Desired[cur].To().ID())
			res[next] = append(res[next], p)
		}
	}

	return res
}
