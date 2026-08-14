package xoxtable

import "testing"

func TestIsGameFinish(t *testing.T) {
	tests := []struct {
		name string
		tbl  Table
		want bool
	}{
		{"bos tahta", Table{}, false},
		{"kismen dolu", Table{X, O, E, E, X, E, E, E, E}, false},
		{"berabere ile tam dolu", Table{X, O, X, X, O, O, O, X, X}, true},
		{"kazanan ile tam dolu", Table{X, X, X, O, O, X, O, X, O}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGameFinish(tt.tbl); got != tt.want {
				t.Errorf("IsGameFinish(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGameWin(t *testing.T) {
	tests := []struct {
		name string
		tbl  Table
		want bool
	}{
		{"bos tahta", Table{}, false},
		{"satir 0 kazanir", Table{X, X, X, O, O, E, E, E, E}, true},
		{"satir 1 kazanir", Table{O, E, E, X, X, X, O, E, E}, true},

		// Regression: the old code checked t[0]!=Empty on the last row (instead of t[6]).
		// Since t[0] was empty, a real win was incorrectly returned as "false".
		{"satir 2 kazanir (t0 bosken regression)", Table{E, E, E, E, E, E, X, X, X}, true},

		{"sutun 0 kazanir", Table{X, O, O, X, E, E, X, E, E}, true},
		{"sutun 1 kazanir", Table{O, X, O, E, X, E, E, X, E}, true},
		{"sutun 2 kazanir", Table{O, O, X, E, E, X, E, E, X}, true},
		{"ana kosegen kazanir", Table{X, O, O, O, X, E, E, E, X}, true},
		{"ters kosegen kazanir", Table{O, O, X, O, X, E, X, E, E}, true},

		{"kazanan yok, kismen dolu", Table{X, O, E, E, X, E, E, E, O}, false},

		// Regression: the old code incorrectly returned "true" when the bottom row
		// was entirely empty (Empty==Empty==Empty) while t[0] was filled.
		{"sahte kazanma korumasi: bos alt satir + dolu t0", Table{X, E, E, E, E, E, E, E, E}, false},

		{"berabere, tam dolu, kazanan yok", Table{X, O, X, X, O, O, O, X, X}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGameWin(tt.tbl); got != tt.want {
				t.Errorf("IsGameWin(%v): got %v, want %v", tt.tbl, got, tt.want)
			}
		})
	}
}
