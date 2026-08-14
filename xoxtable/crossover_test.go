package xoxtable

import "testing"

func TestRecord_AccumulatesOutcomes(t *testing.T) {
	node := &MovementTree{}

	node.Record(3)
	node.Record(-5)
	node.Record(1)

	if node.Visits() != 3 {
		t.Errorf("Visits() = %d, want 3", node.Visits())
	}
	if node.Score() != -1 {
		t.Errorf("Score() = %d, want -1", node.Score())
	}
	if got := node.quality(); got != -1.0/3.0 {
		t.Errorf("quality() = %v, want %v", got, -1.0/3.0)
	}
}

func TestQuality_UntriedMoveIsNeutral(t *testing.T) {
	if got := (&MovementTree{}).quality(); got != 0 {
		t.Errorf("denenmemis hamlenin quality degeri = %v, want 0", got)
	}
}

func TestPreferred_PicksTheBetterRecord(t *testing.T) {
	better := &MovementTree{score: 30, visits: 10} // ortalama 3
	worse := &MovementTree{score: 10, visits: 10}  // ortalama 1

	// Deterministic regardless of argument order, and not a coin flip.
	for range 50 {
		if preferred(worse, better) != better {
			t.Fatal("preferred: daha iyi sicilli dugum secilmedi")
		}
		if preferred(better, worse) != better {
			t.Fatal("preferred: sira degisince daha iyi sicilli dugum secilmedi")
		}
	}
}

func TestPreferred_LosingMoveGivesWayToAnUntriedOne(t *testing.T) {
	losing := &MovementTree{score: -20, visits: 10} // ortalama -2
	untried := &MovementTree{}

	for range 50 {
		if preferred(losing, untried) != untried {
			t.Fatal("preferred: kaybettirdigi bilinen hamle denenmemise tercih edildi")
		}
	}
}

func TestPreferred_ProvenMoveIsKeptOverAnUntriedOne(t *testing.T) {
	winning := &MovementTree{score: 20, visits: 10} // ortalama 2
	untried := &MovementTree{}

	for range 50 {
		if preferred(winning, untried) != winning {
			t.Fatal("preferred: kazandirdigi bilinen hamle denenmemis ugruna birakildi")
		}
	}
}

func TestPreferred_NoEvidenceEitherWayUsesBothParents(t *testing.T) {
	a := &MovementTree{}
	b := &MovementTree{}

	seen := map[*MovementTree]bool{}
	for range 200 {
		seen[preferred(a, b)] = true
	}
	if len(seen) != 2 {
		t.Error("preferred: kanit yokken secim yazi-tura olmali, tek tarafa sabitlenmis")
	}
}

func TestCrossover_CarriesTheRecordHalvedAndResetsOnMutation(t *testing.T) {
	build := func(score, visits int32) *MovementTree {
		root := NewMovementTree(X)
		root.ObtainMovement(Table{})
		child := root.nexts[0]
		child.score, child.visits = score, visits
		return root
	}

	// No mutation: the record follows the inherited move, halved.
	parent := build(40, 20)
	child := Crossover(parent, parent, 0)
	if got := child.nexts[0]; got.Score() != 20 || got.Visits() != 10 {
		t.Errorf("sicil yarilanarak tasinmadi: score=%d visits=%d, want 20/10", got.Score(), got.Visits())
	}

	// Always mutating: the record described the move that was replaced.
	mutated := Crossover(build(40, 20), build(40, 20), 1)
	if got := mutated.nexts[0]; got.Visits() != 0 || got.Score() != 0 {
		t.Errorf("mutasyondan sonra sicil sifirlanmadi: score=%d visits=%d", got.Score(), got.Visits())
	}
}
