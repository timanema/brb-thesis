package algo

//TODO: redo this to actually compare edges

// Needs to use RoutingTable pre-filtering!
func FindDependants(r RoutingTable) map[uint64]map[uint64]struct{} {
	res := make(map[uint64]map[uint64]struct{})

	for dst, paths := range r {
		deps := make(map[uint64]struct{})

		for _, p := range paths {
			for _, e := range p.P[1:] {
				deps[uint64(e.From().ID())] = struct{}{}
			}
		}

		res[dst] = deps
	}

	return res
}

// TODO: make better later
func decideDeadlock(cur, dep uint64) bool {
	return cur < dep
}

// Should be used after RoutingTable has been filtered
func FixDeadlocks(r RoutingTable, deps map[uint64]map[uint64]struct{}) {
	for dst, paths := range r {
		for i, p := range paths {
			if p.Prio {
				continue
			}

			for _, e := range p.P[1:] {
				from := uint64(e.From().ID())

				_, dep := deps[dst][from]
				_, mutual := deps[from][dst]

				// Check for conflict
				if dep && mutual {
					//fmt.Printf("conflict in %v (to %v over %v): %v\n", p, dst, from, decideDeadlock(dst, from))

					// If conflict, make priority path
					if decideDeadlock(dst, from) {
						r[dst][i].Prio = true
						break
					}
				}
			}
		}
	}
}
