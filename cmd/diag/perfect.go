package main

import (
	"math/rand/v2"

	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

// A perfect X is the only benchmark that cannot be gamed by training: perfect
// play by both sides is a draw, so an O that never loses to this opponent is
// playing the game correctly, whatever it was trained against. Measuring only
// against random opponents would flatter an agent trained against random
// opponents, which is exactly the comparison being made here.

var memo = map[[10]byte]int{}

func key(table game.Table, turn game.Cell) [10]byte {
	var k [10]byte
	for i, cell := range table {
		k[i] = byte(cell)
	}
	k[9] = byte(turn)
	return k
}

func other(turn game.Cell) game.Cell {
	if turn == game.X {
		return game.O
	}
	return game.X
}

// value is the game-theoretic value of table for X, with turn to move: +1 if
// X can force a win, -1 if O can, 0 for a draw.
func value(table game.Table, turn game.Cell) int {
	if game.IsGameWin(table) {
		// Whoever moved last won, and that is the side not to move now.
		if turn == game.X {
			return -1
		}
		return 1
	}
	if game.IsGameFinish(table) {
		return 0
	}

	k := key(table, turn)
	if cached, ok := memo[k]; ok {
		return cached
	}

	best := -2
	if turn == game.O {
		best = 2
	}
	for i := range 9 {
		if table.GetCell(i) != game.E {
			continue
		}
		next := table
		next.SetCell(i, turn)

		v := value(next, other(turn))
		if turn == game.X && v > best {
			best = v
		}
		if turn == game.O && v < best {
			best = v
		}
	}

	memo[k] = best
	return best
}

// perfectMove picks uniformly among X's optimal moves, so a perfect opponent
// still varies its play instead of replaying one game every time.
func perfectMove(table game.Table) int {
	best, moves := -2, []int(nil)
	for i := range 9 {
		if table.GetCell(i) != game.E {
			continue
		}
		next := table
		next.SetCell(i, game.X)

		switch v := value(next, game.O); {
		case v > best:
			best, moves = v, []int{i}
		case v == best:
			moves = append(moves, i)
		}
	}
	return moves[rand.IntN(len(moves))]
}
