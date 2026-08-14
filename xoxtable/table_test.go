package xoxtable

import "testing"

func TestTable_SetCell(t *testing.T) {
	tbl := Table{}
	tbl.SetCell(0, X)
	if tbl[0] != X {
		t.Fatalf("SetCell(0, X): got %v, want %v", tbl[0], X)
	}

	// out-of-range index should be silently ignored, not panic
	tbl.SetCell(-1, O)
	tbl.SetCell(9, O)
	if tbl[0] != X {
		t.Fatalf("out-of-range SetCell tabloyu degistirdi: %v", tbl)
	}
}

func TestTable_GetCell(t *testing.T) {
	tbl := Table{}
	tbl[3] = O

	if got := tbl.GetCell(3); got != O {
		t.Fatalf("GetCell(3): got %v, want %v", got, O)
	}
	if got := tbl.GetCell(-1); got != E {
		t.Fatalf("GetCell(-1): got %v, want Empty", got)
	}
	if got := tbl.GetCell(9); got != E {
		t.Fatalf("GetCell(9): got %v, want Empty", got)
	}
}

func TestTable_IsFull(t *testing.T) {
	tests := []struct {
		name string
		tbl  Table
		want bool
	}{
		{"bos tahta", Table{}, false},
		{"kismen dolu", Table{X, O, X}, false},
		{"tam dolu", Table{X, O, X, O, X, O, O, X, O}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tbl.IsFull(); got != tt.want {
				t.Errorf("IsFull(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		tbl  Table
		want bool
	}{
		{"bos tahta", Table{}, true},
		{"tek hucre dolu", Table{E, E, X}, false},
		{"tam dolu", Table{X, O, X, O, X, O, O, X, O}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tbl.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_IsEqual(t *testing.T) {
	a := Table{X, O, E, E, X, E, E, E, O}
	b := a
	c := Table{X, O, E, E, O, E, E, E, O}

	if !a.IsEqual(b) {
		t.Errorf("IsEqual: ayni tablolar farkli raporlandi")
	}
	if a.IsEqual(c) {
		t.Errorf("IsEqual: farkli tablolar ayni raporlandi")
	}
}

func TestTable_Reset(t *testing.T) {
	tbl := Table{X, O, X, O, X, O, X, O, X}
	tbl.Reset()
	if !tbl.IsEmpty() {
		t.Errorf("Reset sonrasi tablo bos degil: %v", tbl)
	}
}

func TestTable_String(t *testing.T) {
	tbl := Table{X, O, E, E, X, E, E, E, O}
	want := "XO.\n.X.\n..O\n"
	if got := tbl.String(); got != want {
		t.Errorf("String(): got %q, want %q", got, want)
	}
}
