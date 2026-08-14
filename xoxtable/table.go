package xoxtable

// We need minimized table for representation of the game map, so we can use it for genetic algorithm

import "strings"

type Cell byte

const X Cell = 'X'
const O Cell = 'O'
const E Cell = 0x00

// We can represent the game map as a 1D array of 9 cells, where each cell can be X, O, or Empty
// Maximum size is 9*8byte = 72byte, which is very small and can be easily stored in memory
type Table [9]Cell

func (t *Table) SetCell(index int, value Cell) {
	if index < 0 || index >= len(t) {
		return
	}
	t[index] = value
}

func (t *Table) GetCell(index int) Cell {
	if index < 0 || index >= len(t) {
		return E
	}
	return t[index]
}

func (t *Table) IsFull() bool {
	for _, cell := range t {
		if cell == E {
			return false
		}
	}
	return true
}

func (t *Table) IsEmpty() bool {
	for _, cell := range t {
		if cell != E {
			return false
		}
	}
	return true
}

func (t *Table) IsEqual(table Table) bool {
	return *t == table
}

func (t *Table) Reset() {
	for i := range t {
		t[i] = E
	}
}

func (t *Table) String() string {
	var result strings.Builder
	for i, cell := range t {
		if cell == E {
			result.WriteString(".")
		} else {
			result.WriteString(string(cell))
		}
		if (i+1)%3 == 0 {
			result.WriteString("\n")
		}
	}
	return result.String()
}
