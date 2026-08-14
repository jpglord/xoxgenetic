package xoxtable

import "testing"

func TestTransforms_AreEightDistinctPermutations(t *testing.T) {
	seen := map[[9]int]bool{}
	for _, p := range transforms {
		seen[p] = true

		hit := map[int]bool{}
		for _, index := range p {
			if index < 0 || index > 8 {
				t.Fatalf("transform index disarida: %v", p)
			}
			hit[index] = true
		}
		if len(hit) != 9 {
			t.Fatalf("transform bir permutasyon degil: %v", p)
		}
	}
	if len(seen) != 8 {
		t.Fatalf("8 farkli simetri bekleniyordu, %d bulundu", len(seen))
	}
}

func TestCanonical_IsStableAcrossEverySymmetry(t *testing.T) {
	board := Table{X, O, E, E, X, E, E, E, O}

	want, _ := canonical(board)
	for k, p := range transforms {
		got, _ := canonical(applyTransform(board, p))
		if got != want {
			t.Errorf("simetri %d icin kanonik form farkli: got %v, want %v", k, got, want)
		}
	}
}

func TestCanonical_MovesMapBackToCallerOrientation(t *testing.T) {
	// A corner opening: whichever corner it is, the canonical form is the
	// same, and translating a canonical move back must land on a cell that
	// is empty in the caller's own orientation.
	for _, corner := range []int{0, 2, 6, 8} {
		board := Table{}
		board.SetCell(corner, X)

		canon, index := canonical(board)
		for move := range 9 {
			if canon.GetCell(move) != E {
				continue
			}
			actual := transforms[index][move]
			if board.GetCell(actual) != E {
				t.Errorf("kose %d: kanonik hamle %d -> %d dolu hucreye dustu", corner, move, actual)
			}
		}
	}
}

func TestCanonical_CornersShareOneClassButEdgesDoNot(t *testing.T) {
	classOf := func(cell int) Table {
		board := Table{}
		board.SetCell(cell, X)
		canon, _ := canonical(board)
		return canon
	}

	corner := classOf(0)
	for _, other := range []int{2, 6, 8} {
		if classOf(other) != corner {
			t.Errorf("kose %d, kose 0 ile ayni sinifa dusmeli", other)
		}
	}
	if classOf(1) == corner {
		t.Error("kenar acilisi kose acilisiyla ayni sinifa dusmemeli")
	}
	if classOf(4) == corner {
		t.Error("merkez acilisi kose acilisiyla ayni sinifa dusmemeli")
	}
}
