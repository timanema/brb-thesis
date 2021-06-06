package algo

import (
	"gonum.org/v1/gonum/graph"
)

type conflict struct {
	path    Path
	overlap []graph.WeightedEdge
	a, b    uint64
}

func findOverlap(p1, p2 Path) [][]graph.WeightedEdge {
	var res [][]graph.WeightedEdge
	var cur []graph.WeightedEdge
	overlap := false

	for _, e := range p1.P {
		for i := len(p2.P) - 1; i >= 0; i-- {
			re := p2.P[i]
			if e.From().ID() == re.To().ID() && e.To().ID() == re.From().ID() {
				// e == reverse of other e
				overlap = true
				cur = append(cur, e)
			} else if overlap {
				// Previously overlap, no longer overlap so commit current overlap
				overlap = false
				res = append(res, cur)
				cur = nil
			}
		}
	}

	return res
}

func findConflicts(original Path, paths []Path) []conflict {
	var res []conflict

	for _, p := range paths {
		overlap := findOverlap(original, p)
		if overlap != nil {
			for _, o := range overlap {
				res = append(res, conflict{
					path:    p,
					overlap: o,
					a:       uint64(o[0].From().ID()),
					b:       uint64(o[len(o)-1].To().ID()),
				})
			}
		}
	}

	return res
}

// TODO: make better later
func decideDeadlock(p Path, c conflict) bool {
	if c.a == c.b {
		panic("invalid conflict")
	}

	if p.Prio || c.path.Prio {
		return false
	}

	return c.a < c.b
}

func FixDeadlocks(r RoutingTable) {
	for dst, paths := range r {
		for i, p := range paths {
			for ndst, npaths := range r {
				if dst == ndst {
					continue
				}

				for _, c := range findConflicts(p, npaths) {
					if decideDeadlock(p, c) {
						r[dst][i].Prio = true
					}
				}
			}
		}
	}
}

// TODO: fix this function, doesn't resolve all deadlocks
// Should be used after RoutingTable has been filtered
//func FixDeadlocks(r RoutingTable, deps map[uint64]map[uint64]struct{}) {
//	for dst, paths := range r {
//		for i, p := range paths {
//			if p.Prio {
//				continue
//			}
//
//			for _, e := range p.P[1:] {
//				from := uint64(e.From().ID())
//
//				_, dep := deps[dst][from]
//				_, mutual := deps[from][dst]
//
//				// Check for conflict
//				if dep && mutual {
//					//fmt.Printf("conflict in %v (to %v over %v): %v\n", p, dst, from, decideDeadlock(dst, from))
//
//					// If conflict, make priority path
//					if decideDeadlock(dst, from) {
//						r[dst][i].Prio = true
//						break
//					}
//				}
//			}
//		}
//	}
//}
