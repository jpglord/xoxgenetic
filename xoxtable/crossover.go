package xoxtable

import "math/rand/v2"

// Crossover breeds a child MovementTree from two parents of the same side
// (agentType). The tree itself is the genome: nodes are matched by the
// board state they were reached at (current), since that's the locus two
// independently-grown trees share. Where both parents explored a state,
// the child inherits one parent's move at random. Where only one parent
// explored a state, that subtree is cloned. In both cases mutateProb is
// the per-node chance the inherited move is discarded and re-rolled via
// randomMove, giving the population a way to explore states no parent's
// move at that node ever led to.
func Crossover(a, b *MovementTree, mutateProb float64) *MovementTree {
	return crossoverNode(a, b, mutateProb, true)
}

// crossoverNode merges a and b, which (when both non-nil) represent the
// same board state for the same side. At least one of a, b must be non-nil.
func crossoverNode(a, b *MovementTree, mutateProb float64, isRoot bool) *MovementTree {
	primary, other := a, b
	if primary == nil {
		primary, other = b, a
	}

	source := preferred(primary, other)
	move := source.move

	// The record is evidence about this move, so it travels with it, halved
	// so recent games outweigh old ones. Without that decay a long-lived
	// node accumulates so many visits that new evidence can no longer move
	// its average, which matters here because the opponent population keeps
	// changing underneath it.
	score, visits := source.score/2, source.visits/2

	if rand.Float64() < mutateProb {
		move = randomMove(primary.current)
		score, visits = 0, 0 // the record described the move just replaced
	}

	result := primary.current
	result.SetCell(int(move), primary.agentType)
	won := IsGameWin(result)

	child := &MovementTree{
		current:   primary.current,
		move:      move,
		agentType: primary.agentType,
		won:       won,
		finished:  won || IsGameFinish(result),
		score:     score,
		visits:    visits,
	}
	child.nexts = child.mergeNexts(a, b, mutateProb, isRoot)
	return child
}

// preferred picks whose answer the child inherits at this position. Judging
// the move by its own record is the whole point: an agent's score judges its
// entire tree at once, so a move that loses games rides along on the strength
// of everything else in that tree and selection never reaches it.
//
// A move with no record counts as neutral rather than as a special case,
// which is what balances exploring against exploiting: a move already known
// to do better than nothing is kept over an untried one, while a move known
// to lose gives way to something not yet tried.
func preferred(primary, other *MovementTree) *MovementTree {
	if other == nil {
		return primary
	}

	switch p, o := primary.quality(), other.quality(); {
	case o > p:
		return other
	case p > o:
		return primary
	}

	if rand.Float64() < 0.5 {
		return other
	}
	return primary
}

// mergeNexts unions the explored children of a and b, matching entries by
// board state (current) so shared states get crossed over and states only
// one parent explored are cloned, then keeps only the children that are
// still reachable from mtree in play.
//
// The filter matters far more than it looks: the child inherits one
// parent's move at this node but both parents' subtrees below it, and the
// subtree the other parent grew under its own different move can never be
// visited again. Without pruning, those dead nodes are copied into every
// descendant forever - measured at ~99% of all nodes by generation 60,
// which is what exhausted memory during long training runs.
func (mtree *MovementTree) mergeNexts(a, b *MovementTree, mutateProb float64, isRoot bool) []*MovementTree {
	var aNexts, bNexts []*MovementTree
	if a != nil {
		aNexts = a.nexts
	}
	if b != nil {
		bNexts = b.nexts
	}

	matchedB := make([]bool, len(bNexts))
	merged := make([]*MovementTree, 0, len(aNexts)+len(bNexts))

	for _, an := range aNexts {
		var match *MovementTree
		for j, bn := range bNexts {
			if !matchedB[j] && bn.current.IsEqual(an.current) {
				match = bn
				matchedB[j] = true
				break
			}
		}
		if !mtree.reaches(an.current, isRoot) {
			continue
		}
		merged = append(merged, crossoverNode(an, match, mutateProb, false))
	}
	for j, bn := range bNexts {
		if matchedB[j] || !mtree.reaches(bn.current, isRoot) {
			continue
		}
		merged = append(merged, crossoverNode(nil, bn, mutateProb, false))
	}

	return merged
}

// reaches reports whether a child node responding to board state child can
// actually be arrived at from mtree during play. Every state here is
// canonical, so this asks whether some opponent reply leads to child's
// symmetry class rather than comparing boards cell by cell.
//
// The root is a dummy container whose own move is never played: an agent
// starts a game by calling root.ObtainMovement on the board it first faces.
// For X that is the empty board itself, so X's first real node responds to
// root.current unchanged; O only moves after X has opened, so O's first
// real node responds to a board with one opponent cell on it. Every node
// below the root is reached on the agent's following turn, after one
// opponent reply to mtree's result - and nothing follows a node whose
// result already ended the game.
func (mtree *MovementTree) reaches(child Table, isRoot bool) bool {
	opponent := X
	if mtree.agentType == X {
		opponent = O
	}

	if isRoot {
		if mtree.agentType == X {
			return mtree.current.IsEqual(child)
		}
		return someReplyLeadsTo(mtree.current, child, opponent)
	}
	if mtree.finished {
		return false
	}
	return someReplyLeadsTo(mtree.Result(), child, opponent)
}

// someReplyLeadsTo reports whether the opponent can play one cell on base
// and land in child's symmetry class.
func someReplyLeadsTo(base, child Table, opponent Cell) bool {
	for i, cell := range base {
		if cell != E {
			continue
		}
		reply := base
		reply.SetCell(i, opponent)
		if reached, _ := canonical(reply); reached.IsEqual(child) {
			return true
		}
	}
	return false
}
