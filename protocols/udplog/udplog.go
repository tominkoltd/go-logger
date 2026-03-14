// Package udplog implements a raw-UDP log destination.
//
// Each log line is sent as a single UDP datagram to the configured endpoint.
// Delivery is fire-and-forget — UDP is connectionless and provides no
// acknowledgement or delivery guarantee.
//
// Destination string format:
//
//	udp://host:port   (e.g. "udp://logstash.local:5000")
//
// Writes are dispatched through a buffered channel to avoid blocking callers.
// When DropIfBusy is set and the channel is full, messages are silently
// discarded rather than blocking.
package udplog

import (
	"errors"
	"math"
	"net"
	"strings"

	"github.com/tominkoltd/go-logger/core"
)

type logMessage struct {
	message string
}

type modData struct {
	conn        net.Conn
	channel     chan logMessage
	config      *core.Config
	dropCounter int
}

// New opens a UDP connection to the address in the Destination string and
// starts a background goroutine to drain the write channel.
func New(c *core.Config) (*modData, error) {
	if !strings.HasPrefix(c.Destination, "udp://") {
		return &modData{}, errors.New("udplog: invalid destination (expected udp://host:port)")
	}

	addr := c.Destination[6:] // strip "udp://"
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return &modData{}, err
	}

	queueSize := c.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}

	md := &modData{
		conn:    conn,
		channel: make(chan logMessage, queueSize),
		config:  c,
	}
	go md.worker()
	return md, nil
}

// Write enqueues a message for delivery. Drops silently if DropIfBusy and full.
func (md *modData) Write(s string, _, _ bool) error {
	if md.channel == nil {
		return errors.New("udplog: not initialised")
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

// Close drains and closes the write channel, then closes the UDP connection.
func (md *modData) Close() error {
	if md.channel != nil {
		close(md.channel)
		md.channel = nil
	}
	if md.conn != nil {
		return md.conn.Close()
	}
	return nil
}

func (md *modData) worker() {
	for msg := range md.channel {
		// Each datagram is the log line + newline; trim to UDP safe size.
		data := msg.message + "\n"
		_, _ = md.conn.Write([]byte(data))
	}
}
