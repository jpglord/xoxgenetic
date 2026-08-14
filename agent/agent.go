package agent

import (
	"github.com/google/uuid"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// Agent plays one side (X or O) of XOX, caching its response to every board
// state it has seen in a movement tree that persists across rounds.
type Agent interface {
	New(id uuid.UUID, generation uint32) Agent
	// NewRound resets the agent's cursor to the root of its movement tree
	// so a fresh game starts from the empty board, while keeping every
	// state cached from previous rounds.
	NewRound()
	Play(table game.Table) game.Decision
	GetUUID() uuid.UUID
	GetRootTree() *game.MovementTree
	GetScore() int32
	AddScore(int32)
	ResetScore()
	// Reward credits the outcome of the round just played to every node the
	// agent actually used, so a move that loses can be selected against even
	// when the agent carrying it scored well overall.
	Reward(outcome int32)
	IsWin() bool
	IsFinish() bool
}
