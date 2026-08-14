package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Akif-jpg/xoxgenetic/agent"
	"github.com/Akif-jpg/xoxgenetic/gaengine"
	game "github.com/Akif-jpg/xoxgenetic/xoxtable"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== XOX: Genetic Tic-Tac-Toe ===")
	fmt.Println("Cells are numbered 0-8, left to right, top to bottom:")
	fmt.Println("0 1 2")
	fmt.Println("3 4 5")
	fmt.Println("6 7 8")

	fmt.Print("\nChoose mode: [1] You vs Agent  [2] Agent vs Agent  [3] Train via Genetic Algorithm: ")
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		fmt.Println("\nInput closed, exiting.")
		return
	}

	switch strings.TrimSpace(line) {
	case "2":
		runAgentVsAgent(reader)
		return
	case "3":
		runGeneticTraining(reader)
		return
	}
	runHumanVsAgent(reader)
}

// warmupGenerations is how long both teams are trained against each other
// before O switches to random opponents. Training only against a co-evolving
// rival caps how good O gets: measured over three runs each at 100
// generations, co-evolution alone still lost 0-3.8% of games to perfect play,
// while finishing against random opponents lost none at all. The warm-up is
// kept because it is what makes the X team worth anything, and because
// starting from a population that already plays competently costs nothing.
const warmupGenerations = 100

// runGeneticTraining trains in two phases and then lets the human play X
// against the best-scoring trained O agent. After each game the user can
// train more generations (continuing from the current population, not
// restarting) before playing again, or quit.
func runGeneticTraining(reader *bufio.Reader) {
	cfg := gaengine.DefaultConfig()

	fmt.Println("\n=== Genetic Algorithm Training ===")
	fmt.Printf("Population: %d per side, %d rounds/generation, elite %.0f%%, survivors %.0f%%, mutation %.0f%%\n",
		cfg.PopulationSize, cfg.RoundsPerGeneration, cfg.EliteFraction*100, cfg.SurvivorFraction*100, cfg.MutationRate*100)
	fmt.Printf("Phase 1: %d generations of X vs O. Phase 2: O alone vs random opponents.\n", warmupGenerations)

	logger, err := newTrainingLogger()
	if err != nil {
		fmt.Printf("Warning: could not open training log file (%v); continuing without logging.\n", err)
	} else {
		defer logger.close()
	}

	fmt.Print("How many generations for phase 2? [default 100]: ")
	n := readIntOrDefault(reader, 100)

	gen := gaengine.NewGeneration(cfg)
	gen = trainGenerations(gen, cfg, warmupGenerations, logger, coevolution)
	gen = trainGenerations(gen, cfg, n, logger, versusRandom)

	for {
		best := gaengine.Best(gen.OTeam)
		fmt.Printf("\nPlaying against the best O agent from generation %d (score %d)\n", gen.Number, best.GetScore())
		playHumanVsTrainedAgent(reader, best)

		fmt.Print("\n[Enter] play again  [t] train more generations  [q] quit: ")
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("\nInput closed, exiting.")
			return
		}
		switch strings.TrimSpace(strings.ToLower(line)) {
		case "q":
			fmt.Println("Goodbye!")
			return
		case "t":
			fmt.Print("How many more generations? [default 10]: ")
			more := readIntOrDefault(reader, 10)
			gen = trainGenerations(gen, cfg, more, logger, versusRandom)
		}
	}
}

// Training phases. coevolution trains both teams against each other;
// versusRandom trains only O, against random opponents.
const (
	coevolution  = "coevolution"
	versusRandom = "vs-random"
)

// trainGenerations evolves gen forward n times in the given phase, printing
// each evaluated generation's best/avg score as progress and appending the
// same stats to logger if it's non-nil. Stats must be read from gen *after*
// evolving: a generation plays its rounds (mutating its agents' scores in
// place) before building and returning the next one, so gen only holds real
// scores once that has run.
func trainGenerations(gen *gaengine.Generation, cfg gaengine.Config, n int, logger *trainingLogger, phase string) *gaengine.Generation {
	for range n {
		var next *gaengine.Generation
		if phase == coevolution {
			next = gen.Evolve(cfg)
		} else {
			next = gen.EvolveAgainstRandom(cfg)
		}

		stats := gen.Stats()
		if phase == coevolution {
			fmt.Printf("[%s] Generation %d -> %d: best X=%d (avg %.2f), best O=%d (avg %.2f)\n",
				phase, gen.Number, next.Number, stats.X.Best, stats.X.Avg, stats.O.Best, stats.O.Avg)
		} else {
			// The X team is idle in this phase, so its scores are stale.
			fmt.Printf("[%s] Generation %d -> %d: best O=%d (avg %.2f)\n",
				phase, gen.Number, next.Number, stats.O.Best, stats.O.Avg)
		}
		if logger != nil {
			logger.log(stats, phase)
		}
		gen = next
	}
	return gen
}

// readIntOrDefault reads one line and parses it as a positive int, falling
// back to def on empty input or a parse error.
func readIntOrDefault(reader *bufio.Reader, def int) int {
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		fmt.Println("\nInput closed, exiting.")
		os.Exit(0)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	n, convErr := strconv.Atoi(line)
	if convErr != nil || n <= 0 {
		fmt.Printf("Invalid number, using default (%d).\n", def)
		return def
	}
	return n
}

// playHumanVsTrainedAgent plays one game with the human as X against a
// trained O agent, resetting the agent's cursor to its root first so it
// doesn't pick up mid-tree from a previous game.
func playHumanVsTrainedAgent(reader *bufio.Reader, oAgent agent.Agent) {
	oAgent.NewRound()
	var table game.Table

	fmt.Println("\n--- New game (you are X) ---")
	printBoard(table)

	for {
		move := readHumanMove(reader, table)
		table.SetCell(move, game.X)
		printBoard(table)

		if game.IsGameWin(table) {
			fmt.Println("You win!")
			return
		}
		if game.IsGameFinish(table) {
			fmt.Println("It's a draw!")
			return
		}

		node := oAgent.Play(table)
		table = node.Result()
		fmt.Printf("Agent plays %d\n", node.Move())
		printBoard(table)

		if node.Won() {
			fmt.Println("Agent wins!")
			return
		}
		if node.Finished() {
			fmt.Println("It's a draw!")
			return
		}
	}
}

// runHumanVsAgent runs repeated You-vs-Agent rounds, reusing the agent's
// movement tree (and cached responses) across rounds.
func runHumanVsAgent(reader *bufio.Reader) {
	// The tree lives outside the game loop so it keeps every board state the
	// agent has seen and its cached response, across restarts.
	root := game.NewMovementTree(game.O)

	for {
		playGame(reader, root)

		fmt.Println()
		fmt.Println("Movement tree explored so far:")
		printMovementTree(root, 0)

		fmt.Print("\nPlay again? [Enter = yes, q = quit]: ")
		line, err := reader.ReadString('\n')
		if err == io.EOF || strings.TrimSpace(strings.ToLower(line)) == "q" {
			fmt.Println("Goodbye!")
			return
		}
	}
}

// runAgentVsAgent runs repeated CPU vs CPU rounds between a fresh X agent
// and O agent, reusing each agent's movement tree (and cached responses)
// across rounds.
func runAgentVsAgent(reader *bufio.Reader) {
	xAgent := agent.NewAgent(game.X, 0)
	oAgent := agent.NewAgent(game.O, 0)

	for {
		xAgent.NewRound()
		oAgent.NewRound()
		playAgentGame(xAgent, oAgent)

		fmt.Println("\nAgent X movement tree explored so far:")
		printMovementTree(xAgent.GetRootTree(), 0)
		fmt.Println("\nAgent O movement tree explored so far:")
		printMovementTree(oAgent.GetRootTree(), 0)

		fmt.Print("\nPlay again? [Enter = yes, q = quit]: ")
		line, err := reader.ReadString('\n')
		if err == io.EOF || strings.TrimSpace(strings.ToLower(line)) == "q" {
			fmt.Println("Goodbye!")
			return
		}
	}
}

// playAgentGame runs a single CPU vs CPU round to completion, X moving
// first.
func playAgentGame(xAgent, oAgent agent.Agent) {
	var table game.Table

	fmt.Println("\n--- New game ---")
	printBoard(table)

	for {
		xNode := xAgent.Play(table)
		table = xNode.Result()
		fmt.Printf("Agent X plays %d\n", xNode.Move())
		printBoard(table)
		if xNode.Won() {
			fmt.Println("Agent X wins!")
			return
		}
		if xNode.Finished() {
			fmt.Println("It's a draw!")
			return
		}

		oNode := oAgent.Play(table)
		table = oNode.Result()
		fmt.Printf("Agent O plays %d\n", oNode.Move())
		printBoard(table)
		if oNode.Won() {
			fmt.Println("Agent O wins!")
			return
		}
		if oNode.Finished() {
			fmt.Println("It's a draw!")
			return
		}
	}
}

// playGame runs a single round to completion (win or draw), restarting the
// board but reusing the agent's movement tree for cached responses.
func playGame(reader *bufio.Reader, root *game.MovementTree) {
	var table game.Table
	node := root

	fmt.Println("\n--- New game ---")
	printBoard(table)

	for {
		move := readHumanMove(reader, table)
		table.SetCell(move, game.X)
		printBoard(table)

		if game.IsGameWin(table) {
			fmt.Println("You win!")
			return
		}
		if game.IsGameFinish(table) {
			fmt.Println("It's a draw!")
			return
		}

		decision := node.ObtainMovement(table)
		node = decision.Tree()
		table = decision.Result()
		fmt.Printf("Agent plays %d\n", decision.Move())
		printBoard(table)

		if decision.Won() {
			fmt.Println("Agent wins!")
			return
		}
		if decision.Finished() {
			fmt.Println("It's a draw!")
			return
		}
	}
}

func readHumanMove(reader *bufio.Reader, table game.Table) int {
	for {
		fmt.Print("Your move (0-8): ")
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("\nInput closed, exiting.")
			os.Exit(0)
		}
		n, convErr := strconv.Atoi(strings.TrimSpace(line))
		if convErr != nil || n < 0 || n > 8 {
			fmt.Println("Enter a number between 0 and 8.")
			continue
		}
		if table.GetCell(n) != game.E {
			fmt.Println("That cell is already taken.")
			continue
		}
		return n
	}
}

func printBoard(table game.Table) {
	fmt.Print(table.String())
}

// printMovementTree recursively prints every board state the agent has
// cached a response for, indented by depth.
func printMovementTree(node *game.MovementTree, depth int) {
	indent := strings.Repeat("  ", depth)
	label := "root"
	status := ""
	if depth > 0 {
		label = fmt.Sprintf("move %d", node.Move())
		if node.Won() {
			status = " [WIN]"
		} else if node.Finished() {
			status = " [DRAW]"
		}
	}
	fmt.Printf("%s- %s%s\n", indent, label, status)
	for _, next := range node.Nexts() {
		printMovementTree(next, depth+1)
	}
}
