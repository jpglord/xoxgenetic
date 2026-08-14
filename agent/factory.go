package agent

import (
	"github.com/google/uuid"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// NewAgent builds a fresh agent for the given side (game.X or game.O) and
// generation, with a new UUID and an empty movement tree.
func NewAgent(side game.Cell, generation uint32) Agent {
	switch side {
	case game.X:
		return NewXAgent(uuid.New(), generation)
	case game.O:
		return NewOAgent(uuid.New(), generation)
	default:
		panic("agent: unknown side")
	}
}

// NewAgentFromTree builds a fresh agent for the given side, wrapping an
// already-built movement tree (a crossover child or an unchanged elite)
// with a new UUID, the given generation, and score 0.
func NewAgentFromTree(side game.Cell, generation uint32, tree *game.MovementTree) Agent {
	switch side {
	case game.X:
		return NewXAgentFromTree(uuid.New(), generation, tree)
	case game.O:
		return NewOAgentFromTree(uuid.New(), generation, tree)
	default:
		panic("agent: unknown side")
	}
}
