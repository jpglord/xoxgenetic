package xoxtable

import "math/rand/v2"

// For make a move on gamemap
type movement uint8

// MovementTree is kept deliberately small: a population holds millions of
// these, so every field costs megabytes. result is derived by Result()
// rather than stored, since it is just current with move applied.
//
// current and move are held in canonical form (see symmetry.go), so one
// node answers for every board symmetric to it. Callers therefore go
// through ObtainMovement, which returns a Decision already translated back
// into their own orientation; the accessors below expose the raw canonical
// genome and are meant for inspecting or breeding a tree, not for play.
type MovementTree struct {
	current   Table           // Current status in the game, canonical
	move      movement        // 0-8 represent movement in the game, canonical
	agentType Cell            // X or O to write index at the move
	won       bool            // whether Result() is a win for agentType
	finished  bool            // whether Result() finishes the game (win or draw)
	score     int32           // summed outcome of the games this move took part in
	visits    int32           // how many games that is
	nexts     []*MovementTree // Referance to nexts
}

// Decision is an agent's answer for one concrete board: the genome node
// that supplied it, plus the move and resulting board mapped out of
// canonical space into the orientation the caller asked about. Won and
// Finished need no mapping, since a win is a win in every symmetry.
type Decision struct {
	tree   *MovementTree
	move   int
	result Table
}

// Tree returns the genome node this decision came from, which is what an
// agent advances its cursor to.
func (d Decision) Tree() *MovementTree { return d.tree }

// Move returns the cell index (0-8) to play, in the caller's orientation.
func (d Decision) Move() int { return d.move }

// Result returns the board the caller asked about with Move() applied.
func (d Decision) Result() Table { return d.result }

// Won reports whether Result() wins for the agent that made the move.
func (d Decision) Won() bool { return d.tree.won }

// Finished reports whether Result() ends the game (win or draw).
func (d Decision) Finished() bool { return d.tree.finished }

// NewMovementTree creates a new MovementTree for the root node (empty board).
func NewMovementTree(agentType Cell) *MovementTree {
	return newMovementNode(Table{}, agentType)
}

// newMovementNode creates a single node for a canonical board state. move is
// chosen randomly, then that move is applied to current with agentType to
// obtain the result; won and finished are computed once here and cached on
// the node, so callers don't need to call IsGameWin/IsGameFinish again from
// outside.
func newMovementNode(current Table, agentType Cell) *MovementTree {
	move := randomMove(current)

	result := current
	result.SetCell(int(move), agentType)

	won := IsGameWin(result)

	return &MovementTree{
		current:   current,
		move:      move,
		agentType: agentType,
		won:       won,
		finished:  won || IsGameFinish(result),
	}
}

func randomMove(table Table) movement {
	emptyCells := make([]movement, 0, 9)
	for i, v := range table {
		if v == E {
			emptyCells = append(emptyCells, movement(i))
		}
	}
	return emptyCells[rand.Int()%len(emptyCells)]
}

// decide translates mtree's canonical answer back into the orientation of
// board, which must be the symmetry of mtree.current selected by p.
func (mtree *MovementTree) decide(board Table, p [9]int) Decision {
	move := p[mtree.move]

	result := board
	result.SetCell(move, mtree.agentType)

	return Decision{tree: mtree, move: move, result: result}
}

// ObtainMovement answers for table: it reduces table to its canonical form,
// reuses the child node holding that form if there is one, and otherwise
// grows a new node with a random move. Because the lookup is canonical, a
// response learned in one orientation is reused in all eight.
func (mtree *MovementTree) ObtainMovement(table Table) Decision {
	canon, index := canonical(table)

	for _, next := range mtree.nexts {
		if next.current.IsEqual(canon) {
			return next.decide(table, transforms[index])
		}
	}

	newTree := newMovementNode(canon, mtree.agentType)
	mtree.nexts = append(mtree.nexts, newTree)
	return newTree.decide(table, transforms[index])
}

// Move returns the canonical cell index (0-8) to be played from this node.
func (mtree *MovementTree) Move() int {
	return int(mtree.move)
}

// Current returns the canonical board state this node responds to.
func (mtree *MovementTree) Current() Table {
	return mtree.current
}

// Result returns the canonical current with Move() applied using agentType.
// It is recomputed on each call rather than stored, to keep the node small.
func (mtree *MovementTree) Result() Table {
	result := mtree.current
	result.SetCell(int(mtree.move), mtree.agentType)
	return result
}

// Won returns whether Result() is a win for agentType.
func (mtree *MovementTree) Won() bool {
	return mtree.won
}

// Finished returns whether Result() ends the game (win or draw).
func (mtree *MovementTree) Finished() bool {
	return mtree.finished
}

// Record credits a finished game's outcome to this node, which took part in
// it. This is what lets selection reach an individual move: an agent's own
// score judges its whole tree at once, so a move that loses games survives
// as long as the rest of the tree carries it. Every node used in a game gets
// the same outcome, which is noisy per game but unbiased over many of them.
func (mtree *MovementTree) Record(outcome int32) {
	mtree.score += outcome
	mtree.visits++
}

// Score returns the summed outcome of the games this node's move took part in.
func (mtree *MovementTree) Score() int32 {
	return mtree.score
}

// Visits returns how many games have been credited to this node.
func (mtree *MovementTree) Visits() int32 {
	return mtree.visits
}

// quality is the node's mean outcome, with an untried move counting as
// neutral so it can still be compared against one that has a record.
func (mtree *MovementTree) quality() float64 {
	if mtree.visits == 0 {
		return 0
	}
	return float64(mtree.score) / float64(mtree.visits)
}

// Nexts returns the child nodes explored so far from this node.
func (mtree *MovementTree) Nexts() []*MovementTree {
	return mtree.nexts
}

// Size returns the number of nodes in the subtree rooted at mtree
// (including mtree itself).
func (mtree *MovementTree) Size() int {
	size := 1
	for _, n := range mtree.nexts {
		size += n.Size()
	}
	return size
}
