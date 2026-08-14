package gaengine

// Config tunes one run of the generational loop.
type Config struct {
	// PopulationSize is the number of X agents and the number of O agents
	// in every generation (team sizes are kept equal).
	PopulationSize int
	// RoundsPerGeneration is how many full rounds (each a reshuffled 1:1
	// X-vs-O pairing, played to completion) are used to score a generation
	// before selection, to reduce noise from any single pairing.
	RoundsPerGeneration int
	// EliteFraction is the top slice of each team (by score) whose trees
	// are carried into the next generation unchanged, wrapped in a fresh
	// Agent with score reset to 0.
	EliteFraction float64
	// SurvivorFraction is the top slice of each team (by score) eligible
	// to breed. It must be at least large enough to give EliteFraction
	// something to draw from, and at least 2 to form breeding pairs.
	SurvivorFraction float64
	// MutationRate is the per-node probability, during crossover, that a
	// child's inherited move is discarded and re-rolled at random instead.
	MutationRate float64
}

// DefaultConfig returns reasonable starting values for tic-tac-toe-scale
// populations.
func DefaultConfig() Config {
	return Config{
		PopulationSize:      1000,
		RoundsPerGeneration: 1000,
		EliteFraction:       0.05,
		SurvivorFraction:    0.5,
		MutationRate:        0.05,
	}
}
