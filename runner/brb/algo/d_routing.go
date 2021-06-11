package algo

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"rp-runner/graphs"
)

type Path struct {
	P    graphs.Path
	Prio bool
}
type RoutingTable map[uint64][]Path
type BroadcastPlan map[uint64][]Path
type NextHopPlan map[uint64][]DolevPath

func (n NextHopPlan) Paths() []DolevPath {
	var res []DolevPath

	for _, p := range n {
		res = append(res, p...)
	}

	return res
}

type DolevPath struct {
	Desired, Actual graphs.Path
	Prio            bool
}

func (p DolevPath) SizeOf() uintptr {
	return p.Desired.SizeOf() + p.Actual.SizeOf() + reflect.TypeOf(p.Prio).Size()
}

func SizeOfMultiplePaths(paths []DolevPath) uintptr {
	res := uintptr(0)

	for _, p := range paths {
		res += p.SizeOf()
	}

	return res
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

func CombineDolevPaths(paths []DolevPath) NextHopPlan {
	res := make(map[uint64][]DolevPath)

	for _, p := range paths {
		if cur := len(p.Actual); len(p.Desired) > cur {
			next := uint64(p.Desired[cur].To().ID())
			res[next] = append(res[next], p)
		}
	}

	return res
}

func AddNextHopEntries(next, additional NextHopPlan) {
	for dst, p := range additional {
		next[dst] = append(next[dst], p...)
	}
}

func SetPiggybacks(next, to NextHopPlan, buf []DolevPath) ([]DolevPath, int) {
	if len(buf) == 0 {
		return buf, 0
	}

	lifts := 0
	newBuf := make([]DolevPath, len(buf))
	copy(newBuf, buf)

	for dst := range next {
		r := make([]DolevPath, 0, len(newBuf))
		for _, b := range newBuf {
			if cur := len(b.Actual); len(b.Desired) > cur {
				n := uint64(b.Desired[cur].To().ID())

				if dst == n {
					to[n] = append(to[n], b)
					lifts += 1
				} else {
					r = append(r, b)
				}
			}
		}

		newBuf = r
	}

	return newBuf, lifts
}

func AddPiggybacks(next NextHopPlan, buf []DolevPath) ([]DolevPath, int) {
	return SetPiggybacks(next, next, buf)
}
