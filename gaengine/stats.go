package gaengine

import "github.com/Akif-jpg/xoxgenetic/agent"

// TeamStats summarizes one team's fitness within a generation.
type TeamStats struct {
	Best int32
	Avg  float64
}

func computeTeamStats(team []agent.Agent) TeamStats {
	best := team[0].GetScore()
	var sum int64
	for _, a := range team {
		s := a.GetScore()
		sum += int64(s)
		if s > best {
			best = s
		}
	}
	return TeamStats{Best: best, Avg: float64(sum) / float64(len(team))}
}

// GenerationStats summarizes both teams' fitness for one generation.
type GenerationStats struct {
	Generation uint32
	X          TeamStats
	O          TeamStats
}

// Stats computes fitness stats for g's current population. Call it before
// Evolve, since Evolve's returned generation always starts at score 0.
func (g *Generation) Stats() GenerationStats {
	return GenerationStats{
		Generation: g.Number,
		X:          computeTeamStats(g.XTeam),
		O:          computeTeamStats(g.OTeam),
	}
}
