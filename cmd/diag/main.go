package main

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/jpglord/xoxgenetic/agent"
	"github.com/jpglord/xoxgenetic/gaengine"
	"github.com/jpglord/xoxgenetic/roundcontroller"
	game "github.com/jpglord/xoxgenetic/xoxtable"
)

// winningCells lists the cells where side would immediately win by playing.
func winningCells(table game.Table, side game.Cell) []int {
	var cells []int
	for i := range 9 {
		if table.GetCell(i) != game.E {
			continue
		}
		trial := table
		trial.SetCell(i, side)
		if game.IsGameWin(trial) {
			cells = append(cells, i)
		}
	}
	return cells
}

type result struct {
	win, draw, loss           int
	winChances, winsTaken     int
	blockChances, blocksTaken int
}

// evaluate plays O against the given X and, beyond the raw outcome, checks the
// two tactics that matter most: converting an immediate win, and blocking the
// opponent's immediate win.
func evaluate(o agent.Agent, games int, xMove func(game.Table) int) result {
	var r result
	for range games {
		o.NewRound()
		table := game.Table{}
		for {
			if table.IsFull() {
				r.draw++
				break
			}
			table.SetCell(xMove(table), game.X)
			if game.IsGameWin(table) {
				r.loss++
				break
			}
			if game.IsGameFinish(table) {
				r.draw++
				break
			}

			wins := winningCells(table, game.O)
			threats := winningCells(table, game.X)

			decision := o.Play(table)
			move := decision.Move()

			if len(wins) > 0 {
				r.winChances++
				if slices.Contains(wins, move) {
					r.winsTaken++
				}
			} else if len(threats) > 0 {
				// Only counted when there is nothing better to do than defend.
				r.blockChances++
				if slices.Contains(threats, move) {
					r.blocksTaken++
				}
			}

			table = decision.Result()
			if decision.Won() {
				r.win++
				break
			}
			if decision.Finished() {
				r.draw++
				break
			}
		}
	}
	return r
}

func randomMove(table game.Table) int {
	empties := make([]int, 0, 9)
	for i := range 9 {
		if table.GetCell(i) == game.E {
			empties = append(empties, i)
		}
	}
	return empties[rand.IntN(len(empties))]
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return 100 * float64(a) / float64(b)
}

// clone deep-copies a tree so evaluation never grows the population's own
// genome: crossing a tree with itself at zero mutation reproduces it exactly.
func clone(tree *game.MovementTree) agent.Agent {
	return agent.NewAgentFromTree(game.O, 0, game.Crossover(tree, tree, 0))
}

func report(label string, tree *game.MovementTree) {
	r := evaluate(clone(tree), 2000, randomMove)
	p := evaluate(clone(tree), 500, perfectMove)

	fmt.Printf("%-22s | rastgele X: %5.1f%% kazanc %5.1f%% berabere %5.1f%% KAYIP  (kazanc firsati %4.1f%%, blok %4.1f%%) | mukemmel X: %5.1f%% KAYIP %5.1f%% berabere\n",
		label,
		pct(r.win, r.win+r.draw+r.loss), pct(r.draw, r.win+r.draw+r.loss), pct(r.loss, r.win+r.draw+r.loss),
		pct(r.winsTaken, r.winChances), pct(r.blocksTaken, r.blockChances),
		pct(p.loss, p.win+p.draw+p.loss), pct(p.draw, p.win+p.draw+p.loss))
}

// train runs coevolution generations of team-against-team play, then random
// generations where only O is trained, against random opponents.
func train(cfg gaengine.Config, coevolution, random int) *game.MovementTree {
	gen := gaengine.NewGeneration(cfg)

	for i := range coevolution {
		next := gen.Evolve(cfg)
		if i == coevolution-1 && random == 0 {
			return gaengine.Best(gen.OTeam).GetRootTree()
		}
		gen = next
	}
	for i := range random {
		next := gen.EvolveAgainstRandom(cfg)
		if i == random-1 {
			return gaengine.Best(gen.OTeam).GetRootTree()
		}
		gen = next
	}
	return gaengine.Best(gen.OTeam).GetRootTree()
}

func main() {
	cfg := gaengine.DefaultConfig()
	fmt.Printf("pop=%d rounds/gen=%d mutation=%.0f%% | odul WIN=%d DRAW=%d LOOSE=%d\n\n",
		cfg.PopulationSize, cfg.RoundsPerGeneration, cfg.MutationRate*100,
		roundcontroller.WIN_SCORE, roundcontroller.DRAW_SCORE, roundcontroller.LOOSE_SCORE)

	const repeats = 3
	modes := []struct {
		name                string
		coevolution, random int
	}{
		{"co-evolution 100", 100, 0},
		{"rastgele 100", 0, 100},
		{"hibrit 100+100", 100, 100},
	}

	for _, m := range modes {
		for r := range repeats {
			report(fmt.Sprintf("%s #%d", m.name, r+1), train(cfg, m.coevolution, m.random))
		}
		fmt.Println()
	}
}
