package graphs

import (
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"math"
	"sort"
)

type cost struct {
	id   uint64
	cost int
}

// Bellman-Ford (needs to be capable of handling negative weights)
func BellmanFord(s int64, edges AdjacencyMap, additionalWeight [][]int) []int64 {
	queue := make([]int64, 0)
	inQ := make([]bool, len(edges))

	// Step 1: Init graph
	distance := make([]float64, len(edges))
	predecessor := make([]int64, len(edges))

	for i := 0; i < len(distance); i++ {
		distance[i] = math.MaxInt64
		predecessor[i] = -1
	}

	distance[s] = 0
	queue = append(queue, s)
	inQ[s] = true

	// Step 2: Relax edges repeatedly
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		inQ[n] = false

		for t, e := range edges[n] {
			if e == nil {
				continue
			}

			add := 0.0
			if len(additionalWeight) > 0 {
				add = float64(additionalWeight[n][t])
			}

			if d := distance[n] + e.Weight() + add; d < distance[t] {
				distance[t] = d
				predecessor[t] = n

				if !inQ[t] {
					queue = append(queue, int64(t))
				}
			}
		}
	}

	return predecessor
}

func ShortestPath(s, t int64, edges AdjacencyMap, additionalWeight [][]int) (Path, error) {
	// Use BellmanFord to get predecessor map
	predecessor := BellmanFord(s, edges, additionalWeight)

	// Step 4: build path to target
	res := make([]graph.WeightedEdge, 0)
	cur := t

	for cur != s {
		next := predecessor[cur]

		if next == -1 {
			return nil, errors.New("no path")
		}

		res = append(res, edges[next][cur])
		cur = next
	}

	return res, nil
}

func walkPath(pred []int64, s, t int64, dist []int) (int, error) {
	cnt := 0
	cur := t

	for cur != s {
		if dist[cur] != -1 {
			cnt += dist[cur]
			break
		}

		next := pred[cur]

		if next == -1 {
			return 0, errors.New("no path")
		}

		cnt += 1
		cur = next
	}

	return cnt, nil
}

func ClosestNodes(s int64, edges AdjacencyMap, amount int) []uint64 {
	// Use BellmanFord to get predecessor map
	predecessor := BellmanFord(s, edges, nil)
	costs := make([]cost, 0, len(edges))
	res := make([]uint64, 0, amount)

	// Used for memoization
	dist := make([]int, len(edges))
	for i := 0; i < len(dist); i++ {
		dist[i] = -1
	}

	for nid := int64(0); nid < int64(len(predecessor)); nid++ {
		c, err := walkPath(predecessor, s, nid, dist)
		if err != nil {
			continue
		}

		dist[nid] = c
		costs = append(costs, cost{
			id:   uint64(nid),
			cost: c,
		})
	}

	// Then use a stable (!!) sort on cost
	sort.SliceStable(costs, func(i, j int) bool {
		return costs[i].cost < costs[j].cost
	})

	if len(costs) < amount {
		panic("not enough paths through the network to find required amount of closest neighbours")
	}

	for i := 0; i < amount; i++ {
		res = append(res, costs[i].id)
	}

	return res
}
