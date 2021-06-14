package algo

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"rp-runner/graphs"
	"sync"
)

type Path struct {
	P    graphs.Path
	Prio bool
}
type RoutingTable map[uint64][]Path
type BroadcastPlan map[uint64][]Path
type NextHopPlan map[uint64][]DolevPath

type FullRoutingTable struct {
	Plan   map[uint64]BroadcastPlan
	BDPlan map[uint64]BrachaDolevRoutingTable
	sync.RWMutex
}

// Only used to ease concurrent write, as the underlying map is directly accessible for concurrent read
func (t *FullRoutingTable) Update(origin uint64, p BroadcastPlan) {
	t.Lock()
	defer t.Unlock()
	t.Plan[origin] = p
}

func (t *FullRoutingTable) UpdateBD(origin uint64, p BrachaDolevRoutingTable) {
	t.Lock()
	defer t.Unlock()
	t.BDPlan[origin] = p
}

func (t *FullRoutingTable) FindMatches(origin, partialId uint64, p graphs.Path, partial bool) []DolevPath {
	r := t.Plan[origin]

	if partial {
		r = t.BDPlan[origin][partialId]
	}

	var res []DolevPath

	for dst, dolevs := range r {
		if dst != uint64(p[0].To().ID()) {
			continue
		}

		for _, d := range dolevs {
			if graphs.IsSubPath(p, d.P) {
				dp := DolevPath{
					Desired: make(graphs.Path, len(d.P)),
					Actual:  p,
					Prio:    d.Prio,
				}

				copy(dp.Desired, d.P)
				res = append(res, dp)
			}
		}
	}

	return res
}

func dolevPathContained(xs []DolevPath, p DolevPath) bool {
	for _, d := range xs {
		if d.Prio == p.Prio && graphs.IsEqualPath(p.Desired, d.Desired) {
			return true
		}
	}

	return false
}

func (t *FullRoutingTable) FindAllMatches(origin, partialId uint64, p []DolevPath, partial bool, except []DolevPath) []DolevPath {
	var res []DolevPath

	for _, d := range p {
		matches := t.FindMatches(origin, partialId, d.Actual, partial)

		for _, m := range matches {
			if !dolevPathContained(append(res, except...), m) {
				res = append(res, m)
				break
			}
		}
	}

	return res
}

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

func CleanDesiredPaths(paths []DolevPath) []DolevPath {
	res := make([]DolevPath, 0, len(paths))
	for _, p := range paths {
		p.Desired = nil
		res = append(res, p)
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

func BuildFullRoutingTable(g *simple.WeightedUndirectedGraph, w, n, f, k int, singleHopNeighbour, combineNext, filterSubpath, bd bool) (*FullRoutingTable, error) {
	nodes := g.Nodes()
	ft := &FullRoutingTable{
		Plan:   make(map[uint64]BroadcastPlan),
		BDPlan: make(map[uint64]BrachaDolevRoutingTable),
	}

	errGr, _ := errgroup.WithContext(context.TODO())

	for nodes.Next() {
		node := nodes.Node()

		errGr.Go(func() error {
			r, err := BuildRoutingTable(g, node, k, w, singleHopNeighbour)
			if err != nil {
				return errors.Wrap(err, "failed to build routing table")
			}

			broadcast, partial := Routing(r, uint64(node.ID()), g, w, n, f,
				singleHopNeighbour, combineNext, filterSubpath, bd)

			ft.Update(uint64(node.ID()), broadcast)
			ft.UpdateBD(uint64(node.ID()), partial)
			return nil
		})
	}

	if err := errGr.Wait(); err != nil {
		return nil, err
	}

	return ft, nil
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
