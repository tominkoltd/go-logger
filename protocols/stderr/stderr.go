// Package stderr implements a log destination that writes to os.Stderr.
//
// All Logger instances that target stderr share a single buffered channel and
// worker goroutine, started automatically when the package is imported.
package stderr

import (
	"os"

	"github.com/tominkoltd/go-logger/core"
)

// logMessage is the unit passed through the write channel to the worker goroutine.
type logMessage struct {
	message string
}

// logChannel is the shared write queue for all stderr-bound log entries.
var logChannel = make(chan logMessage, 50)

func init() {
	go loggerWorker()
}

// modData is the (empty) per-instance handle returned by New.
type modData struct{}

// New returns a new stderr destination handle. The config is accepted for
// interface consistency but is not used — all formatting is applied upstream.
func New(c *core.Config) (*modData, error) {
	return &modData{}, nil
}

// Write enqueues a message for the worker goroutine to print to stderr.
func (md *modData) Write(s string, _, _ bool) error {
	logChannel <- logMessage{message: s}
	return nil
}

// Close is a no-op for the stderr destination. The shared worker goroutine
// and channel are process-scoped and are not closed on a per-instance basis.
func (md *modData) Close() error {
	return nil
}

// loggerWorker drains the message channel and writes each entry to stderr.
func loggerWorker() {
	for log := range logChannel {
		_, _ = os.Stderr.WriteString(log.message)
		_, _ = os.Stderr.WriteString("\n")
	}
}
