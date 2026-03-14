// Package file implements a file-based log destination.
//
// Multiple Logger instances that share the same file path will share a single
// os.File handle and write channel, coordinated via a reference count.
// The underlying goroutine and file handle are released when the last
// consumer calls Close.
package file

import (
	"errors"
	"math"
	"os"
	"sync"

	"github.com/tominkoltd/go-logger/core"
)

// logMessage is the unit passed through the write channel to the worker goroutine.
type logMessage struct {
	message string
}

// modData holds the per-instance state for a file destination.
type modData struct {
	disabled    bool
	fileName    string
	channel     chan logMessage
	config      *core.Config
	dropCounter int // number of messages dropped due to a full queue (DropIfBusy)
}

// Package-level state shared across all file-destination instances.
// Protected by mtx.
var (
	mtx         sync.Mutex
	channels    = map[string]chan logMessage{} // keyed by resolved file path
	connections = map[string]int{}             // reference count per file path
)

// New initialises a file destination for the given config.
// If another Logger is already writing to the same file, the existing channel
// and worker goroutine are reused and the reference count is incremented.
func New(c *core.Config) (*modData, error) {
	md := &modData{config: c}

	// Expect at least "file:" (5 bytes) + one path character.
	if len(c.Destination) < 6 {
		md.disabled = true
		return md, errors.New("invalid file destination")
	}

	fileName := c.Destination[5:] // strip the "file:" prefix
	md.fileName = fileName

	mtx.Lock()
	defer mtx.Unlock()

	ch, ok := channels[fileName]
	if !ok {
		// First consumer for this file — open it and start the worker.
		fh, err := os.OpenFile(
			fileName,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)
		if err != nil {
			md.disabled = true
			return md, err
		}

		ch = make(chan logMessage, c.QueueSize)
		channels[fileName] = ch
		go loggerWorker(fh, ch)
	}

	// Increment the reference count for this file.
	connections[fileName]++

	md.channel = ch
	return md, nil
}

// Write enqueues a message for the worker goroutine to write to the file.
// If DropIfBusy is set and the queue is full, the message is silently
// discarded and the internal drop counter is incremented.
func (md *modData) Write(s string, _, _ bool) error {
	if md.disabled {
		return errors.New("destination disabled")
	}
	if md.channel == nil {
		return errors.New("file channel not initialised")
	}

	if md.config.DropIfBusy {
		select {
		case md.channel <- logMessage{message: s}:
		default:
			if md.dropCounter < math.MaxInt {
				md.dropCounter++
			}
		}
		return nil
	}

	md.channel <- logMessage{message: s}
	return nil
}

// Close decrements the reference count for the destination file.
// When the count reaches zero the write channel is closed, which causes the
// worker goroutine to exit and the file handle to be closed.
func (md *modData) Close() error {
	mtx.Lock()
	defer mtx.Unlock()

	cnt, ok := connections[md.fileName]
	if !ok || cnt <= 0 {
		return nil
	}

	cnt--
	if cnt == 0 {
		// Last consumer — shut down the worker.
		ch, ok := channels[md.fileName]
		if ok && ch != nil {
			close(ch)
		}
		delete(channels, md.fileName)
		delete(connections, md.fileName)
	} else {
		connections[md.fileName] = cnt
	}

	return nil
}

// loggerWorker drains the message channel and writes each entry to the file.
// It exits when the channel is closed (via Close), at which point the deferred
// fileHandler.Close() flushes and releases the OS file handle.
func loggerWorker(fileHandler *os.File, ch <-chan logMessage) {
	defer fileHandler.Close()
	for log := range ch {
		_, _ = fileHandler.WriteString(log.message)
		_, _ = fileHandler.WriteString("\n")
	}
}
