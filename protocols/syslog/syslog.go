//go:build !windows

// Package syslog implements a syslog log destination.
//
// Destination string formats:
//
//	syslog:              – local syslog via /dev/log or /var/run/syslog
//	syslog://host:port   – remote syslog over UDP
//	syslog+tcp://host:port – remote syslog over TCP
//
// Severity mapping:
//
//	Log()   → syslog.Info
//	Warn()  → syslog.Warning
//	Error() → syslog.Err
//
// The SyslogTag config field sets the program identifier visible in syslog
// output (e.g. "myapp" → "myapp[pid]: message"). Defaults to "apm".
//
// Recommended logger.Config for syslog destinations:
//
//	logger.Config{
//	    Destination: "syslog://localhost:514",
//	    ParseColors: true,   // expand ç tokens first …
//	    StripAnsi:   true,   // … then strip ANSI before sending
//	    SyslogTag:   "myapp",
//	}
package syslog

import (
	gosyslog "log/syslog"
	"strings"

	"github.com/tominkoltd/go-logger/core"
)

type modData struct {
	w *gosyslog.Writer
}

// New opens a connection to the syslog daemon described by the Destination string.
func New(c *core.Config) (*modData, error) {
	network, addr := parseAddr(c.Destination)

	tag := c.SyslogTag
	if tag == "" {
		tag = "apm"
	}

	w, err := gosyslog.Dial(network, addr, gosyslog.LOG_INFO|gosyslog.LOG_DAEMON, tag)
	if err != nil {
		return &modData{}, err
	}
	return &modData{w: w}, nil
}

// Write sends the message to syslog at the appropriate severity level.
func (md *modData) Write(s string, isErr, isWarn bool) error {
	if md.w == nil {
		return nil
	}
	switch {
	case isErr:
		return md.w.Err(s)
	case isWarn:
		return md.w.Warning(s)
	default:
		return md.w.Info(s)
	}
}

// Close shuts down the syslog connection.
func (md *modData) Close() error {
	if md.w != nil {
		return md.w.Close()
	}
	return nil
}

// parseAddr extracts the network type and address from a syslog destination string.
//
//	"syslog:"               → ("", "")        local
//	"syslog://host:514"     → ("udp", "host:514")
//	"syslog+tcp://host:514" → ("tcp", "host:514")
func parseAddr(dest string) (network, addr string) {
	switch {
	case strings.HasPrefix(dest, "syslog+tcp://"):
		return "tcp", dest[len("syslog+tcp://"):]
	case strings.HasPrefix(dest, "syslog://"):
		return "udp", dest[len("syslog://"):]
	default:
		return "", "" // local
	}
}
