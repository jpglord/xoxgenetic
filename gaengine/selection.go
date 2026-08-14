package gaengine

import (
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/jpglord/xoxgenetic/agent"
	game "github.com/jpglord/xoxgenetic/xoxtable"
)

// breedWorkers is the fixed number of goroutines that breed children
// concurrently, mirroring roundcontroller's worker-pool sizing.
const breedWorkers = 8

// selectAndBreed produces the next generation's team for one side: the top
// cfg.EliteFraction carry their tree forward unchanged (fresh Agent,
// score 0), and the rest are bred by picking two parents from the top
// cfg.SurvivorFraction via score-weighted roulette selection and crossing
// their trees. Breeding is spread across breedWorkers goroutines: Crossover
// only reads its parent trees (never mutates them), so concurrent workers
// can safely share the same survivor pool without a mutex, and each worker
// writes to its own index of next.
func selectAndBreed(team []agent.Agent, side game.Cell, generation uint32, cfg Config) []agent.Agent {
	n := len(team)

	sorted := make([]agent.Agent, n)
	copy(sorted, team)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetScore() > sorted[j].GetScore()
	})

	eliteCount := int(float64(n) * cfg.EliteFraction)
	survivorCount := max(eliteCount, int(float64(n)*cfg.SurvivorFraction), 2)
	survivorCount = min(survivorCount, n)
	survivors := sorted[:survivorCount]
	weights := rouletteWeights(survivors)

	next := make([]agent.Agent, n)
	for i := range eliteCount {
		next[i] = agent.NewAgentFromTree(side, generation, sorted[i].GetRootTree())
	}

	indices := make(chan int, n-eliteCount)
	for i := eliteCount; i < n; i++ {
		indices <- i
	}
	close(indices)

	var wg sync.WaitGroup
	for range breedWorkers {
		wg.Go(func() {
			for i := range indices {
				parentA := weightedPick(survivors, weights)
				parentB := weightedPick(survivors, weights)
				child := game.Crossover(parentA.GetRootTree(), parentB.GetRootTree(), cfg.MutationRate)
				next[i] = agent.NewAgentFromTree(side, generation, child)
			}
		})
	}
	wg.Wait()

	return next
}

// rouletteWeights turns pool's scores into non-negative breeding weights:
// scores can be negative (LOOSE_SCORE), so weights are shifted by the
// pool's minimum score, and every survivor gets weight >= 1 so no one is
// unbreedable.
func rouletteWeights(pool []agent.Agent) []float64 {
	min := pool[0].GetScore()
	for _, a := range pool[1:] {
		if a.GetScore() < min {
			min = a.GetScore()
		}
	}

	weights := make([]float64, len(pool))
	for i, a := range pool {
		weights[i] = float64(a.GetScore()-min) + 1
	}
	return weights
}

// weightedPick picks one agent from pool with probability proportional to
// weights (roulette-wheel selection).
func weightedPick(pool []agent.Agent, weights []float64) agent.Agent {
	total := 0.0
	for _, w := range weights {
		total += w
	}

	r := rand.Float64() * total
	cumulative := 0.0
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return pool[i]
		}
	}
	return pool[len(pool)-1]
}
