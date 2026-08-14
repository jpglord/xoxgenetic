package gaengine

import (
	"sync"

	"github.com/Akif-jpg/xoxgenetic/agent"
	"github.com/Akif-jpg/xoxgenetic/roundcontroller"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// Generation is one batch of X and O populations, all sharing the same
// generation number.
type Generation struct {
	Number uint32
	XTeam  []agent.Agent
	OTeam  []agent.Agent
}

// NewGeneration builds generation 0: fresh agents with empty movement
// trees, per cfg.PopulationSize.
func NewGeneration(cfg Config) *Generation {
	xTeam, oTeam := roundcontroller.NewTeams(cfg.PopulationSize)
	return &Generation{Number: 0, XTeam: xTeam, OTeam: oTeam}
}

// Evolve scores the current generation over cfg.RoundsPerGeneration rounds,
// then produces the next generation via elitism + score-weighted crossover
// breeding. It does not mutate g; g's agents are left with whatever score
// they accumulated, for inspection (e.g. Best) before being discarded.
func (g *Generation) Evolve(cfg Config) *Generation {
	for range cfg.RoundsPerGeneration {
		for _, a := range g.XTeam {
			a.NewRound()
		}
		for _, a := range g.OTeam {
			a.NewRound()
		}
		roundcontroller.PlayRound(g.XTeam, g.OTeam)
	}

	nextGen := g.Number + 1
	var nextX, nextO []agent.Agent
	var wg sync.WaitGroup
	wg.Go(func() {
		nextX = selectAndBreed(g.XTeam, game.X, nextGen, cfg)
	})
	wg.Go(func() {
		nextO = selectAndBreed(g.OTeam, game.O, nextGen, cfg)
	})
	wg.Wait()

	return &Generation{Number: nextGen, XTeam: nextX, OTeam: nextO}
}

// EvolveAgainstRandom scores and breeds only the O team, playing it against
// random opponents rather than against the X team. Co-evolution makes both
// populations chase each other into an ever narrower set of lines, so an
// agent can get better against its current rival while getting worse at the
// game; training against random play keeps the whole board in view. The X
// team is carried through untouched, since it is not being trained here.
func (g *Generation) EvolveAgainstRandom(cfg Config) *Generation {
	for range cfg.RoundsPerGeneration {
		for _, a := range g.OTeam {
			a.NewRound()
		}
		roundcontroller.PlayRoundVsRandom(g.OTeam)
	}

	nextGen := g.Number + 1
	return &Generation{
		Number: nextGen,
		XTeam:  g.XTeam,
		OTeam:  selectAndBreed(g.OTeam, game.O, nextGen, cfg),
	}
}

// Best returns the highest-scoring agent in team. team must be non-empty.
func Best(team []agent.Agent) agent.Agent {
	best := team[0]
	for _, a := range team[1:] {
		if a.GetScore() > best.GetScore() {
			best = a
		}
	}
	return best
}
