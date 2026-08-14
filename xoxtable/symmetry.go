package xoxtable

// A 3x3 board has 8 symmetries (the dihedral group of the square: four
// rotations, each with and without a mirror). Two boards related by one of
// them are the same position as far as play is concerned, so the movement
// tree stores only one canonical representative per class. Every response
// an agent learns is then shared by up to 8 orientations instead of having
// to be rediscovered separately in each, which both shrinks the genome and
// multiplies how often each of its nodes is exercised.

// transforms[t][i] is the index in the untransformed board whose value ends
// up at index i once transform t is applied.
var transforms = buildTransforms()

func buildTransforms() [8][9]int {
	// Cells are numbered left to right, top to bottom, so a quarter turn
	// clockwise moves index 6 to index 0, 3 to 1, and so on.
	rotate := [9]int{6, 3, 0, 7, 4, 1, 8, 5, 2}
	mirror := [9]int{2, 1, 0, 5, 4, 3, 8, 7, 6}

	// compose(p, q) is the permutation of applying p and then q.
	compose := func(p, q [9]int) [9]int {
		var out [9]int
		for i := range out {
			out[i] = p[q[i]]
		}
		return out
	}

	var all [8][9]int
	current := [9]int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	for k := range 4 {
		all[k] = current
		all[k+4] = compose(current, mirror)
		current = compose(current, rotate)
	}
	return all
}

// applyTransform returns t as seen through the given index permutation.
func applyTransform(t Table, p [9]int) Table {
	var out Table
	for i := range out {
		out[i] = t[p[i]]
	}
	return out
}

// canonical returns the lexicographically smallest of t's 8 symmetries,
// along with the transform that produces it. A move chosen on the returned
// board maps back to the caller's orientation through that transform: a
// canonical move m is the caller's cell transforms[t][m], because that is
// the cell whose value was read into position m.
// Canonical exposes canonical for callers that need to reason about the
// symmetry classes a movement tree is keyed by.
func Canonical(t Table) (Table, int) {
	return canonical(t)
}

func canonical(t Table) (Table, int) {
	best, bestIndex := applyTransform(t, transforms[0]), 0
	for k := 1; k < len(transforms); k++ {
		candidate := applyTransform(t, transforms[k])
		if lessTable(candidate, best) {
			best, bestIndex = candidate, k
		}
	}
	return best, bestIndex
}

func lessTable(a, b Table) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}
