# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
go build ./...              # build everything
go vet ./...                # static checks
go test ./...                # run all tests
go test ./xoxtable/...       # run tests for one package
go test ./xoxtable/ -run TestName   # run a single test
go run .                     # play interactively (prompts for mode)
```

There is no linter or CI config in this repo beyond `go vet`.

## Architecture

Module: `github.com/Akif-jpg/xoxgenetic`. Three packages:

- **`xoxtable`** — the game engine. `Table` (`table.go`) is a `[9]Cell` board (`X`, `O`, or `E` for empty); `check.go` has the pure win/finish predicates (`IsGameWin`, `IsGameFinish`). `movement.go` defines `MovementTree`, which is both the CPU's move-selection logic and its memory: each node holds a board state (`current`), a chosen `move` (currently picked at random via `randomMove`), the resulting `Table`, and cached `won`/`finished` flags. `ObtainMovement(table)` walks a node's `nexts` for a child matching `table`; if none exists it creates one via `newMovementNode` and appends it. The tree is never pruned, so it accumulates every board state a given root has ever seen — this is the substrate the eventual genetic algorithm will select/mutate over.

- **`agent`** — wraps a `MovementTree` behind the `Agent` interface (`New`, `NewRound`, `Play`, `GetUUID`, `GetRootTree`). `XAgent`/`OAgent` each keep two pointers into their tree: `root` (fixed, used for inspecting/printing everything explored) and `cursor` (advances via `ObtainMovement` as a game progresses, reset to `root` by `NewRound`). `NewRound` must be called at the start of every game or the cursor will pick up mid-tree from the previous game. `factory.NewAgent(side game.Cell, generation uint32)` is the single entry point for constructing an agent of either side with a fresh UUID; use it rather than calling `NewXAgent`/`NewOAgent` directly outside the package.

- **`main`** — CLI entry point with two modes selected by prompt: `runHumanVsAgent` (a human plays X against a single long-lived `MovementTree` built directly on `game.O`, bypassing the `agent` package) and `runAgentVsAgent` (two `agent.Agent`s, X and O, play each other via `agent.NewAgent`). Each mode's outer loop persists its tree(s)/agent(s) across "play again?" rounds so cached responses accumulate over a session; `printMovementTree` recursively dumps whatever's been explored after each round.

Note the two modes don't share code paths: `runHumanVsAgent` talks to `xoxtable.MovementTree` directly, `runAgentVsAgent` goes through `agent.Agent`. If you change move-selection or caching behavior in `MovementTree`, check whether it needs to change in both.

`generation` fields on `XAgent`/`OAgent` are currently unused — no genetic algorithm (fitness, selection, crossover, mutation) exists yet. That's the natural next phase of this project.
