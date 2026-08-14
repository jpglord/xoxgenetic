package roundcontroller

import (
	"github.com/Akif-jpg/xoxgenetic/agent"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// NewTeams builds size fresh generation-0 X and O agents. Each call returns
// a brand new population, since gaengine needs a distinct team slice per
// generation rather than a single mutated-in-place set of agents.
func NewTeams(size int) (xTeam, oTeam []agent.Agent) {
	xTeam = make([]agent.Agent, size)
	oTeam = make([]agent.Agent, size)

	for i := range size {
		xTeam[i] = agent.NewAgent(game.X, 0)
	}
	for i := range size {
		oTeam[i] = agent.NewAgent(game.O, 0)
	}

	return xTeam, oTeam
}
