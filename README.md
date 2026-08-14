# XOX with Go Genetic

A tic-tac-toe player evolved by a genetic algorithm, written in Go.

There is no minimax here, no neural network, and no hand-written strategy. Agents
start out playing at random, and the only thing that ever improves them is
selection: the ones that score better get to breed, and their offspring inherit
a mixture of their parents' answers.

The end result plays the game **perfectly on defence** — measured against a
minimax opponent, it never loses a single game, which is the theoretical
best a second player can do at tic-tac-toe.

```
                    │  vs random X          │  vs perfect (minimax) X
 training (100 gen) │  loss rate            │  loss rate
────────────────────┼───────────────────────┼─────────────────────────
 co-evolution only  │  0.2%  2.0%  2.2%     │  3.8%  0.0%  3.2%
 random opponents   │  0.0%  0.0%  0.0%     │  0.0%  0.0%  0.0%
 hybrid (default)   │  0.0%  0.0%  0.0%     │  0.0%  0.0%  0.0%
```

*Three independent training runs per row. Against a random opponent the trained
agent also wins ~91% of games, so it is not merely playing safe.*

## Running it

```sh
go run .        # pick a mode: play a human game, watch two agents, or train
go test ./...   # 28 tests
go run ./cmd/diag   # the benchmark harness, including the minimax judge
```

Mode `[3]` trains a population and then puts you at the board against the best
agent it produced. Training writes a `ga_training_<timestamp>.csv` with
per-generation statistics so runs can be compared afterwards.

## The idea: the tree is the genome

An agent's entire strategy is a `MovementTree`. Each node holds a board
position and the move this agent plays there; children are the positions that
can follow. The tree grows lazily — a node is only created the first time an
agent is actually asked about that position — because most agents never reach
most of the game.

That makes the genome an unusual shape for a GA: not a fixed-length vector but
a sparse, growing lookup table. Crossover therefore works structurally, matching
two parents' nodes by the position they answer and inheriting one parent's move
at each. Mutation re-rolls a single node's move.

```
xoxtable/   the game, and the genome: Table, win/draw predicates,
            MovementTree, crossover, the symmetry layer
agent/      an Agent wraps a tree, tracks a cursor through one game,
            and carries the score selection ranks it by
roundcontroller/  plays matches on a worker pool and settles the results
gaengine/   the genetic algorithm: config, generations, selection, breeding
cmd/diag/   benchmarks, including a memoised minimax "perfect X" judge
```

## How it got here

The interesting part of this project was not the first version — it was
everything that went wrong afterwards, and what measuring it revealed. Almost
every step below started as a confident hypothesis that turned out to be wrong.

### It ran out of memory

Training died around generation 136 on a 32 GB machine. The obvious suspects —
population size, agent UUIDs, the statistics — turned out to account for
kilobytes. Measuring the trees found the real cause: crossover merged both
parents' children, but a subtree grown under the *other* parent's move can never
be reached once a different move is inherited. Those nodes were dead on arrival
and copied into every descendant forever.

**99% of every tree was unreachable.** 8,235 nodes per agent at generation 60,
of which 78 were reachable. Pruning them, plus dropping a never-read field and
deriving one that was being stored, took the node from 64 to 40 bytes and made
memory flat: **200 generations in 1.69 s with a ~15 MB live heap.**

An earlier suggestion — capping tree size — was rejected, correctly, on the
grounds that it would fight selection rather than help it. Removing *unreachable*
nodes is different: they can never be played, so they can never affect fitness.

### It stopped learning at generation 100

The co-evolutionary score kept drifting, which made it look like progress. An
absolute benchmark — best agent against a purely random opponent — showed
otherwise:

```
untrained      64.1% loss
100 gens       47.9% loss
1000 gens      54.0% loss
5000 gens      47.9% loss     ← 5000 generations equals 100
```

Two barriers explained it. **Coverage:** 23.5% of the positions the agent faced
had never been seen, and were answered by a coin flip. **Credit assignment:** the
genome is hundreds of independent decisions, but fitness was a single number
from ten games, so most nodes were never exercised at all in a generation.

### Symmetry: 289 decisions, not thousands

A 3×3 board has eight symmetries. Storing positions in a canonical form means a
lesson learned in one orientation is reused in all eight. Folding them shows how
small the real problem is:

```
distinct positions where O must choose:  289
distinct positions where X must choose:  338
complete games: 255,168 total → 26,830 distinct
```

Unseen positions collapsed from **23.5% to 0.6%**, and the useful training
horizon stretched from ~100 to ~1000 generations. It also reframed everything:
1000 agents were storing overlapping copies of what is really a 289-entry table,
and judging the whole table with one number.

### Evaluate more, not longer

Raising rounds per generation from 10 to 1000 fixed the noise half of credit
assignment by brute force: an agent judged on 1000 games has a score that
reflects its quality rather than its luck in the draw. **100 generations of this
beat 5000 generations of the old configuration.**

### The reward experiment that failed

The trained agent blocked 94% of threats but converted only 61% of its own
immediate wins. The reward table looked like the culprit: with `WIN=3, DRAW=1,
LOOSE=-5`, turning a draw into a win gains `+2` while avoiding a loss gains `+6`,
so defence is worth three times offence.

Raising `WIN` to 6 to close that gap made the agent **worse at everything,
including the metric it was aimed at** — losses 29% vs 15%, wins taken 49% vs
59%, blocking 59% vs 80%. Both sides share the reward table, so a bigger win
bonus mostly makes the *opponent* gamble harder, and neither population ever
consolidates blocking. Blocking is the foundation: an agent that cannot defend
never reaches a position worth winning. The original values were kept.

### Credit where it is actually due

The missed wins were not a matter of preference but of visibility: a move that
settles for a draw instead of winning costs 2 points inside a ~1300-point score.
Invisible. So each node now keeps the record of the games its move took part in,
and crossover picks the parent whose move has the better record instead of
flipping a coin. The record travels with the move, halved each generation so
recent evidence dominates, and resets when the move is mutated. An untried move
counts as neutral, which keeps a proven-good move but lets a proven-losing one
give way to something new.

**Immediate wins taken went from ~58% to 85–96%, losses from ~15% to ~3%.**

### The real ceiling was co-evolution

Even then, the agent still lost a few percent of games to perfect play, and its
blocking had been quietly *degrading* over long runs — from 94% down to 77% by
generation 2000. Training only against a rival that is itself evolving makes both
populations chase each other into an ever narrower set of lines. The agent was
getting better against its opponent while getting worse at the game.

Training O against random opponents instead — and training only O, since the
second player has the harder job — closed it completely: **zero losses against
perfect play, in every run, with none of the run-to-run variance.** The default
is now a hybrid: a co-evolution warm-up so the X team means something, then
random opponents to broaden what O actually sees.

## What we took away from it

- **Co-evolutionary scores cannot measure progress.** They move when the
  opponent moves. Every real conclusion here came from an absolute benchmark,
  and the honest one was the minimax judge, because it cannot be gamed by
  choosing convenient training opponents.
- **Measure before optimising, and be ready to be wrong.** Tree size, mutation
  rate, and reward shaping were each a confident diagnosis that measurement
  refuted. The wrong ones were as informative as the right ones.
- **Fitness has to reach the thing being selected.** Almost every ceiling in
  this project was some version of a bad decision surviving inside a good agent.
- **Exploit the structure of the problem.** Symmetry turned a vague search into
  a 289-entry table. Nothing else gave a comparable return for the effort.

## Who did what

A genuinely joint project.

**Akif-jpg** wrote the original `xoxtable`, `agent` and `roundcontroller` — the
game engine, the movement tree that became the genome, and the match runner —
and made the calls that mattered most at the turning points: rejecting tree-size
capping as fighting selection, asking for crossover to be parallelised the way
rounds already were, raising evaluation to 1000 rounds per generation, spotting
the agent's passivity at the board before any metric caught it, and diagnosing
co-evolution as the real ceiling, which is what finally solved it.

**Claude** (Anthropic's Claude Code) wrote `gaengine` and the benchmark harness,
added the symmetry layer, the reachability pruning, node-level credit
assignment, and the concurrency work, and ran the measurements that decided
which ideas survived — including its own that did not.

## License

MIT — see [LICENSE](LICENSE). Use it for anything, including learning from it,
which is what it was for.
