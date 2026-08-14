package agent

import (
	"github.com/google/uuid"
	game "github.com/jpglord/xoxgenetic/xoxtable"
)

// XAgent plays as X.
type XAgent struct {
	id         uuid.UUID
	generation uint32
	root       *game.MovementTree
	cursor     *game.MovementTree
	score      int32
	path       []*game.MovementTree
}

// NewXAgent creates a fresh X agent with an empty movement tree.
func NewXAgent(id uuid.UUID, generation uint32) *XAgent {
	root := game.NewMovementTree(game.X)
	return &XAgent{id: id, generation: generation, root: root, cursor: root}
}

// NewXAgentFromTree wraps an existing movement tree (e.g. one produced by
// xoxtable.Crossover, or an elite carried over unchanged) in a fresh X
// agent with the given id/generation and score reset to 0.
func NewXAgentFromTree(id uuid.UUID, generation uint32, tree *game.MovementTree) *XAgent {
	return &XAgent{id: id, generation: generation, root: tree, cursor: tree}
}

func (a *XAgent) New(id uuid.UUID, generation uint32) Agent {
	return NewXAgent(id, generation)
}

func (a *XAgent) NewRound() {
	a.cursor = a.root
	a.path = a.path[:0]
}

func (a *XAgent) Play(table game.Table) game.Decision {
	decision := a.cursor.ObtainMovement(table)
	a.cursor = decision.Tree()
	a.path = append(a.path, a.cursor)
	return decision
}

func (a *XAgent) GetUUID() uuid.UUID {
	return a.id
}

func (a *XAgent) GetRootTree() *game.MovementTree {
	return a.root
}

func (a *XAgent) GetScore() int32 {
	return a.score
}

func (a *XAgent) AddScore(diff int32) {
	a.score += diff
}

func (a *XAgent) ResetScore() {
	a.score = 0
}

func (a *XAgent) Reward(outcome int32) {
	for _, node := range a.path {
		node.Record(outcome)
	}
}

func (a *XAgent) IsWin() bool {
	return a.cursor.Won()
}

func (a *XAgent) IsFinish() bool {
	return a.cursor.Finished()
}
