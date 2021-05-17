package graphs

import "gonum.org/v1/gonum/graph"

func EqualEdge(a, b graph.Edge) bool {
	return a.To().ID() == b.To().ID() && a.From().ID() == b.From().ID()
}

func IsSubPath(a, b Path) bool {
	if len(a) > len(b) {
		return false
	}

	for i, e := range a {
		if !EqualEdge(e, b[i]) {
			return false
		}
	}

	return true
}

func FilterSubpaths(paths []Path) []Path {
	res := make([]Path, 0, len(paths))

	for i, p := range paths {
		sub := false
		for j, sp := range paths {
			if i == j {
				continue
			}

			if IsSubPath(p, sp) {
				sub = true
				break
			}
		}

		if !sub {
			res = append(res, p)
		}
	}

	return res
}
