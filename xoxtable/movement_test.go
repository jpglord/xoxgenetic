package xoxtable

import "testing"

func TestNewMovementTree_RootNode(t *testing.T) {
	root := NewMovementTree(X)

	if root == nil {
		t.Fatal("NewMovementTree: kok dugum nil donduruldu")
	}
	if root.current != (Table{}) {
		t.Errorf("NewMovementTree: current bos tahta olmali, got %v", root.current)
	}
	if root.Move() < 0 || root.Move() > 8 {
		t.Fatalf("NewMovementTree: move gecersiz index: %d", root.Move())
	}

	want := Table{}
	want.SetCell(root.Move(), X)
	if root.Result() != want {
		t.Errorf("NewMovementTree: result = %v, want %v", root.Result(), want)
	}

	if root.Won() {
		t.Errorf("NewMovementTree: tek hamleyle Won() true olmamali")
	}
	if root.Finished() {
		t.Errorf("NewMovementTree: tek hamleyle Finished() true olmamali")
	}
}

func TestObtainMovement_ReturnsSameNodeForEqualTable(t *testing.T) {
	root := NewMovementTree(X)

	state := Table{X, E, E, E, E, E, E, E, E}

	first := root.ObtainMovement(state)
	second := root.ObtainMovement(state)

	if first.Tree() != second.Tree() {
		t.Fatalf("ObtainMovement: ayni tablo icin farkli dugumler dondu")
	}
	if len(root.nexts) != 1 {
		t.Fatalf("ObtainMovement: ayni tablo tekrar eklenmis, nexts uzunlugu = %d, want 1", len(root.nexts))
	}
}

func TestObtainMovement_SymmetricTablesShareOneNode(t *testing.T) {
	root := NewMovementTree(X)

	// All four corner openings are the same position up to rotation, so a
	// response learned for one must be reused for the others.
	corners := []Table{
		{X, E, E, E, E, E, E, E, E},
		{E, E, X, E, E, E, E, E, E},
		{E, E, E, E, E, E, X, E, E},
		{E, E, E, E, E, E, E, E, X},
	}

	first := root.ObtainMovement(corners[0]).Tree()
	for _, corner := range corners[1:] {
		if got := root.ObtainMovement(corner).Tree(); got != first {
			t.Errorf("ObtainMovement: %v simetrik kose icin ayri dugum olusturuldu", corner)
		}
	}
	if len(root.nexts) != 1 {
		t.Fatalf("ObtainMovement: 4 kose tek sinifa dusmeli, nexts uzunlugu = %d, want 1", len(root.nexts))
	}
}

func TestObtainMovement_OpeningsCollapseToThreeClasses(t *testing.T) {
	root := NewMovementTree(O)

	// From O's side the nine possible openings are only three distinct
	// positions: corner, edge and centre.
	for cell := range 9 {
		opening := Table{}
		opening.SetCell(cell, X)
		root.ObtainMovement(opening)
	}

	if len(root.nexts) != 3 {
		t.Fatalf("9 acilis 3 simetri sinifina dusmeli, nexts uzunlugu = %d, want 3", len(root.nexts))
	}
}

func TestObtainMovement_NewTableCreatesNewNode(t *testing.T) {
	root := NewMovementTree(X)

	stateA := Table{X, E, E, E, E, E, E, E, E} // corner
	stateB := Table{E, X, E, E, E, E, E, E, E} // edge

	nodeA := root.ObtainMovement(stateA).Tree()
	nodeB := root.ObtainMovement(stateB).Tree()

	if nodeA == nodeB {
		t.Fatalf("ObtainMovement: farkli tablolar icin ayni dugum donduruldu")
	}
	if len(root.nexts) != 2 {
		t.Fatalf("ObtainMovement: nexts uzunlugu = %d, want 2", len(root.nexts))
	}
}

func TestObtainMovement_AnswersInCallerOrientation(t *testing.T) {
	root := NewMovementTree(X)

	// Only index 2 is empty, so whatever the canonical form is, translating
	// the answer back has to land on the caller's own free cell.
	current := Table{X, X, E, O, O, X, X, O, O}
	decision := root.ObtainMovement(current)

	if decision.Move() != 2 {
		t.Fatalf("Move() = %d, want 2 (tek bos hucre)", decision.Move())
	}

	wantResult := Table{X, X, X, O, O, X, X, O, O}
	if decision.Result() != wantResult {
		t.Fatalf("Result() = %v, want %v", decision.Result(), wantResult)
	}
	if !decision.Won() {
		t.Errorf("Won() = false, want true (result satirda X,X,X iceriyor)")
	}
	if !decision.Finished() {
		t.Errorf("Finished() = false, want true (result tam dolu)")
	}
}

func TestMovementTree_NotFinishedWhileGameInProgress(t *testing.T) {
	root := NewMovementTree(X)

	// current has a single X, the remaining 8 cells are empty. Whichever cell
	// is chosen, result will only ever have 2 X's so a win is impossible, and
	// the board isn't full, so the game must not be finished.
	current := Table{X, E, E, E, E, E, E, E, E}
	decision := root.ObtainMovement(current)

	if decision.Won() {
		t.Errorf("Won() = true, want false (sadece 2 X ile kazanma imkansiz)")
	}
	if decision.Finished() {
		t.Errorf("Finished() = true, want false (tahta hala bos hucreler iceriyor)")
	}
}

func TestMovementTree_FinishedWithoutWinIsDraw(t *testing.T) {
	root := NewMovementTree(X)

	// only index 8 is empty in current; filling it results in a draw board.
	current := Table{X, O, X, X, O, O, O, X, E}
	decision := root.ObtainMovement(current)

	if decision.Move() != 8 {
		t.Fatalf("Move() = %d, want 8 (tek bos hucre)", decision.Move())
	}
	if decision.Won() {
		t.Errorf("Won() = true, want false (beraberlik tahtasi)")
	}
	if !decision.Finished() {
		t.Errorf("Finished() = false, want true (tahta dolu)")
	}
}
