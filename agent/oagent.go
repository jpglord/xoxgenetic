package agent

import (
	"github.com/google/uuid"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// OAgent plays as O.
type OAgent struct {
	id         uuid.UUID
	generation uint32
	root       *game.MovementTree
	cursor     *game.MovementTree
	score      int32
	path       []*game.MovementTree
}

// NewOAgent creates a fresh O agent with an empty movement tree.
func NewOAgent(id uuid.UUID, generation uint32) *OAgent {
	root := game.NewMovementTree(game.O)
	return &OAgent{id: id, generation: generation, root: root, cursor: root}
}

// NewOAgentFromTree wraps an existing movement tree (e.g. one produced by
// xoxtable.Crossover, or an elite carried over unchanged) in a fresh O
// agent with the given id/generation and score reset to 0.
func NewOAgentFromTree(id uuid.UUID, generation uint32, tree *game.MovementTree) *OAgent {
	return &OAgent{id: id, generation: generation, root: tree, cursor: tree}
}

func (a *OAgent) New(id uuid.UUID, generation uint32) Agent {
	return NewOAgent(id, generation)
}

func (a *OAgent) NewRound() {
	a.cursor = a.root
	a.path = a.path[:0]
}

func (a *OAgent) Play(table game.Table) game.Decision {
	decision := a.cursor.ObtainMovement(table)
	a.cursor = decision.Tree()
	a.path = append(a.path, a.cursor)
	return decision
}

func (a *OAgent) GetUUID() uuid.UUID {
	return a.id
}

func (a *OAgent) GetRootTree() *game.MovementTree {
	return a.root
}

func (a *OAgent) GetScore() int32 {
	return a.score
}

func (a *OAgent) AddScore(diff int32) {
	a.score += diff
}

func (a *OAgent) ResetScore() {
	a.score = 0
}

func (a *OAgent) Reward(outcome int32) {
	for _, node := range a.path {
		node.Record(outcome)
	}
}

func (a *OAgent) IsWin() bool {
	return a.cursor.Won()
}

func (a *OAgent) IsFinish() bool {
	return a.cursor.Finished()
}
