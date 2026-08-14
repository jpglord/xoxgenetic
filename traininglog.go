package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Akif-jpg/xoxgenetic/gaengine"
)

// trainingLogger appends one row per evaluated generation to a CSV file so
// training history can be analyzed after the fact, not just watched live
// in the console.
type trainingLogger struct {
	file    *os.File
	w       *csv.Writer
	started time.Time
}

// newTrainingLogger creates a new timestamped CSV file, so separate
// training sessions never mix rows together, and writes its header.
func newTrainingLogger() (*trainingLogger, error) {
	path := fmt.Sprintf("ga_training_%s.csv", time.Now().Format("20060102_150405"))
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	if err := w.Write([]string{"generation", "phase", "best_x_score", "avg_x_score", "best_o_score", "avg_o_score", "elapsed_ms"}); err != nil {
		f.Close()
		return nil, err
	}
	w.Flush()

	fmt.Printf("Logging training stats to %s\n", path)
	return &trainingLogger{file: f, w: w, started: time.Now()}, nil
}

// log appends one generation's stats as a row and flushes immediately, so
// the file has every generation seen so far even if training is
// interrupted. phase records which kind of training produced the row: the X
// columns are only meaningful while both teams are being trained.
func (l *trainingLogger) log(stats gaengine.GenerationStats, phase string) {
	l.w.Write([]string{
		strconv.FormatUint(uint64(stats.Generation), 10),
		phase,
		strconv.Itoa(int(stats.X.Best)),
		strconv.FormatFloat(stats.X.Avg, 'f', 2, 64),
		strconv.Itoa(int(stats.O.Best)),
		strconv.FormatFloat(stats.O.Avg, 'f', 2, 64),
		strconv.FormatInt(time.Since(l.started).Milliseconds(), 10),
	})
	l.w.Flush()
}

func (l *trainingLogger) close() {
	l.w.Flush()
	l.file.Close()
}
