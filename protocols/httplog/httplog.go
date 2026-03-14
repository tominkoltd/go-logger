// Package httplog implements an HTTP/HTTPS log destination.
//
// Each log line is sent as an HTTP POST request to the configured endpoint.
// The body is plain text (the formatted log line); Content-Type is text/plain.
//
// Destination string formats:
//
//	http://host/path
//	https://host/path
//
// Writes are dispatched through a buffered channel to avoid blocking callers.
// When DropIfBusy is set and the channel is full, messages are silently
// discarded rather than blocking.
//
// Failed POST requests are silently dropped — the destination is best-effort.
package httplog

import (
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/tominkoltd/go-logger/core"
)

type logMessage struct {
	message string
}

type modData struct {
	endpoint    string
	channel     chan logMessage
	config      *core.Config
	client      *http.Client
	dropCounter int
	stopCh      chan struct{}
}

// New validates the destination URL and starts a background goroutine to drain
// the write channel.
func New(c *core.Config) (*modData, error) {
	if !strings.HasPrefix(c.Destination, "http://") && !strings.HasPrefix(c.Destination, "https://") {
		return &modData{}, errors.New("httplog: invalid destination (expected http:// or https://)")
	}

	queueSize := c.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}

	md := &modData{
		endpoint: c.Destination,
		channel:  make(chan logMessage, queueSize),
		config:   c,
		client:   &http.Client{},
		stopCh:   make(chan struct{}),
	}
	go md.worker()
	return md, nil
}

// Write enqueues a message for delivery. Drops silently if DropIfBusy and full.
func (md *modData) Write(s string, _, _ bool) error {
	if md.channel == nil {
		return errors.New("httplog: not initialised")
	}
	if md.config.DropIfBusy {
		select {
		case md.channel <- logMessage{s}:
		default:
			if md.dropCounter < math.MaxInt {
				md.dropCounter++
			}
		}
		return nil
	}
	md.channel <- logMessage{s}
	return nil
}

// Close drains and closes the write channel, then signals the worker to stop.
func (md *modData) Close() error {
	if md.channel != nil {
		close(md.channel)
		md.channel = nil
	}
	return nil
}

func (md *modData) worker() {
	for msg := range md.channel {
		body := strings.NewReader(msg.message + "\n")
		resp, err := md.client.Post(md.endpoint, "text/plain", body)
		if err == nil {
			resp.Body.Close()
		}
	}
}
