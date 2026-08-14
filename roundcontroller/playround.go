package roundcontroller

import (
	"math/rand/v2"
	"sync"

	"github.com/jpglord/xoxgenetic/agent"
	"github.com/jpglord/xoxgenetic/xoxtable"
)

type Match struct {
	xAgent agent.Agent
	oAgent agent.Agent
}

const numWorkers = 8

func PlayRound(xTeam, oTeam []agent.Agent) {
	length := min(len(xTeam), len(oTeam))
	matchChannel := make(chan Match, length)

	shuffledO := make([]agent.Agent, len(oTeam))
	copy(shuffledO, oTeam)
	rand.Shuffle(len(shuffledO), func(i, j int) {
		shuffledO[i], shuffledO[j] = shuffledO[j], shuffledO[i]
	})

	go func() {
		defer close(matchChannel)
		for i := range length {
			matchChannel <- Match{
				xAgent: xTeam[i],
				oAgent: shuffledO[i],
			}
		}
	}()

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go RoundWorker(&wg, matchChannel)
	}
	wg.Wait()
}

// PlayRoundVsRandom has every O agent play one game against a fresh random
// opponent instead of against the X team. Training a population only against
// its own rival makes it specialise on the narrow set of lines that rival
// happens to play; a random opponent instead spreads play across the whole
// space of positions a careless X can reach.
func PlayRoundVsRandom(oTeam []agent.Agent) {
	jobs := make(chan agent.Agent, len(oTeam))
	go func() {
		defer close(jobs)
		for _, a := range oTeam {
			jobs <- a
		}
	}()

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for o := range jobs {
				playVsRandom(o)
			}
		})
	}
	wg.Wait()
}

func playVsRandom(o agent.Agent) {
	table := xoxtable.Table{}
	for {
		empties := make([]int, 0, 9)
		for i := range 9 {
			if table.GetCell(i) == xoxtable.E {
				empties = append(empties, i)
			}
		}
		if len(empties) == 0 {
			settle(o, DRAW_SCORE)
			return
		}

		table.SetCell(empties[rand.IntN(len(empties))], xoxtable.X)
		if xoxtable.IsGameWin(table) {
			settle(o, LOOSE_SCORE)
			return
		}
		if xoxtable.IsGameFinish(table) {
			settle(o, DRAW_SCORE)
			return
		}

		decision := o.Play(table)
		table = decision.Result()
		if decision.Won() {
			settle(o, WIN_SCORE)
			return
		}
		if decision.Finished() {
			settle(o, DRAW_SCORE)
			return
		}
	}
}

func RoundWorker(wg *sync.WaitGroup, match <-chan Match) {
	defer wg.Done()
	for m := range match {
		table := xoxtable.Table{}
		for {
			table = m.xAgent.Play(table).Result()
			if m.xAgent.IsWin() {
				settle(m.xAgent, WIN_SCORE)
				settle(m.oAgent, LOOSE_SCORE)
				break
			}
			if m.xAgent.IsFinish() {
				settle(m.xAgent, DRAW_SCORE)
				settle(m.oAgent, DRAW_SCORE)
				break
			}
			table = m.oAgent.Play(table).Result()
			if m.oAgent.IsWin() {
				settle(m.oAgent, WIN_SCORE)
				settle(m.xAgent, LOOSE_SCORE)
				break
			}
			if m.oAgent.IsFinish() {
				settle(m.xAgent, DRAW_SCORE)
				settle(m.oAgent, DRAW_SCORE)
				break
			}
		}
	}
}

// settle applies a finished match to one agent, both as the score selection
// ranks it by and as credit on the individual moves it actually played.
func settle(a agent.Agent, outcome int32) {
	a.AddScore(outcome)
	a.Reward(outcome)
}
