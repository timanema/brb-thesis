package algo

import (
	"rp-runner/graphs"
)

type BroadcastPlan map[uint64][]graphs.Path

type DolevPath struct {
	Desired, Actual graphs.Path
}

func combinePaths(paths []graphs.Path) BroadcastPlan {
	res := make(BroadcastPlan)

	for _, p := range paths {
		next := uint64(p[0].To().ID())
		res[next] = append(res[next], p)
	}

	return res
}

func DolevRouting(r graphs.RoutingTable, combine, filter bool) BroadcastPlan {
	br := make([]graphs.Path, 0, len(r))
	for _, g := range r {
		br = append(br, g...)
	}

	if filter {
		br = graphs.FilterSubpaths(br)
	}
	if combine {
		return combinePaths(br)
	}

	res := make(BroadcastPlan)
	for _, p := range br {
		res[uint64(p[0].To().ID())] = append(res[uint64(p[0].To().ID())], p)
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
