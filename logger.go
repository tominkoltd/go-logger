/*
Package logger provides a flexible, multi-destination logging library for Go.

# Overview

A Logger is created with one or more Config entries — each one defines an independent
output destination with its own formatting, filtering, and buffering settings.
Multiple destinations can be active simultaneously; every log call is dispatched
to all of them in order.

# Destinations

	stdout                          – standard output (default when Destination is "")
	stderr                          – standard error
	file:<path>                     – append to a file  (e.g. "file:./app.log")
	syslog:                         – local syslog (/dev/log or /var/run/syslog)
	syslog://host:port              – remote syslog over UDP
	syslog+tcp://host:port          – remote syslog over TCP
	udp://host:port                 – raw UDP datagram per log line
	http://endpoint                 – HTTP POST per log line (async, non-blocking)
	https://endpoint                – same over TLS

# Basic usage

	log, err := logger.New(logger.Config{
		Destination:     "stdout",
		TimeStampFormat: "2006-01-02 15:04:05",
		Prefix:          "[app]",
	})
	if err != nil {
		panic(err)
	}
	defer log.Close()

	log.Log("server started")
	log.Warn("disk usage above 90%%")
	log.Error("connection refused")

	// Optional level and tag arguments (any order):
	log.Log("request handled", 2, "http")   // level=2, tag="http"

# Log levels

Pass an int argument to set the log level (default 0).  Use IgnoreBelow / IgnoreAbove
in Config to filter by level range.

# Tags

Pass a string argument to attach a tag to a message.  Use AllowedTags (comma-separated)
in Config to whitelist specific tags; an empty AllowedTags passes everything.

# Colour syntax

When ParseColors is enabled, inline colour tokens are expanded to ANSI escape codes:

	ç<n>-          foreground colour  (256-colour palette index)
	ç<fg>,<bg>-    foreground + background
	ç<fg>,<bg>,<attr>-  foreground + background + attribute (e.g. "1" for bold)
	çR-            reset to default

The trigger character defaults to 'ç' and can be changed via ParseColorChar.
*/
package logger

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tominkoltd/go-logger/core"
	"github.com/tominkoltd/go-logger/protocols/file"
	"github.com/tominkoltd/go-logger/protocols/httplog"
	"github.com/tominkoltd/go-logger/protocols/stderr"
	"github.com/tominkoltd/go-logger/protocols/stdout"
	"github.com/tominkoltd/go-logger/protocols/syslog"
	"github.com/tominkoltd/go-logger/protocols/udplog"
)

// Logger holds one or more active output configurations.
// Use New or MustNew to construct one.
type Logger struct {
	confs    []*core.Config
	disabled bool
}

// Config describes a single logging destination and its behaviour.
type Config struct {
	// Destination selects the output target.
	// Supported values: "stdout" (default), "stderr", "file:<path>".
	Destination string

	// TimeStampFormat is a Go time-layout string prepended to every message.
	// Leave empty to omit the timestamp.
	TimeStampFormat string

	// QueueSize is the internal channel buffer depth (default 100).
	QueueSize int

	// ReportInterval is reserved for future periodic-stats support.
	ReportInterval int

	// DropIfBusy discards messages silently when the output queue is full
	// instead of blocking the caller.
	DropIfBusy bool

	// Prefix is a static string inserted after the timestamp and before the message.
	Prefix string

	// StripAnsi removes ANSI escape sequences from the final output.
	StripAnsi bool

	// AlignLines indents continuation lines of multi-line messages so they
	// align with the first character of the message body (past the prefix).
	AlignLines bool

	// AlignSeparator, when non-zero, is inserted between the prefix column and
	// the continuation indent (e.g. '|').
	AlignSeparator rune

	// AllowedTags is a comma-separated whitelist of log tags.
	// When non-empty only messages carrying one of the listed tags are written.
	AllowedTags string

	// IgnoreBelow discards messages whose level is strictly below this value.
	// Zero disables the lower bound.
	IgnoreBelow int

	// IgnoreAbove discards messages whose level is strictly above this value.
	// Zero disables the upper bound.
	IgnoreAbove int

	// ParseColors enables inline colour-token expansion (see package doc).
	ParseColors bool

	// ParseColorChar overrides the colour-token trigger character (default 'ç').
	ParseColorChar rune

	// IgnoreLog suppresses plain Log() messages on this destination.
	IgnoreLog bool

	// IgnoreError suppresses Error() messages on this destination.
	IgnoreError bool

	// IgnoreWarn suppresses Warn() messages on this destination.
	IgnoreWarn bool

	// SyslogTag is the program tag used by syslog destinations.
	// Defaults to "apm" when empty. Ignored by non-syslog destinations.
	SyslogTag string
}

// New creates a Logger from one or more Config values.
// All destinations are initialised; the first non-nil error encountered is
// returned alongside the (partially valid) Logger so the caller can decide
// whether to abort or continue with the remaining destinations.
func New(confs ...Config) (*Logger, error) {
	if len(confs) == 0 {
		return nil, errors.New("at least one Config is required")
	}

	var outErr error
	l := &Logger{}

	for _, c := range confs {
		// Apply the queue-size default before copying into the internal config.
		if c.QueueSize == 0 {
			c.QueueSize = 100
		}

		gc := &core.Config{
			Destination:     c.Destination,
			TimeStampFormat: c.TimeStampFormat,
			QueueSize:       c.QueueSize,
			ReportInterval:  c.ReportInterval,
			DropIfBusy:      c.DropIfBusy,
			Prefix:          c.Prefix,
			StripAnsi:       c.StripAnsi,
			AlignLines:      c.AlignLines,
			AlignSeparator:  c.AlignSeparator,
			IgnoreBelow:     c.IgnoreBelow,
			IgnoreAbove:     c.IgnoreAbove,
			ParseColors:     c.ParseColors,
			IgnoreLog:       c.IgnoreLog,
			IgnoreError:     c.IgnoreError,
			IgnoreWarn:      c.IgnoreWarn,
			SyslogTag:       c.SyslogTag,
			AllowedTags:     make(map[string]bool),
		}

		// Resolve the colour-token trigger character.
		if c.ParseColors {
			if c.ParseColorChar != 0 {
				gc.ParseColorChar = string(c.ParseColorChar)
			} else {
				gc.ParseColorChar = "ç"
			}
		}

		// Initialise the output destination module.
		switch {
		case c.Destination == "stdout" || c.Destination == "":
			modLink, err := stdout.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink

		case c.Destination == "stderr":
			modLink, err := stderr.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink

		case strings.HasPrefix(c.Destination, "file:"):
			modLink, err := file.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink

		case c.Destination == "syslog:" ||
			strings.HasPrefix(c.Destination, "syslog://") ||
			strings.HasPrefix(c.Destination, "syslog+tcp://"):
			modLink, err := syslog.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink

		case strings.HasPrefix(c.Destination, "udp://"):
			modLink, err := udplog.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink

		case strings.HasPrefix(c.Destination, "http://") ||
			strings.HasPrefix(c.Destination, "https://"):
			modLink, err := httplog.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink
		}

		// Parse the comma-separated AllowedTags whitelist into a fast lookup map.
		if c.AllowedTags != "" {
			for _, t := range strings.Split(c.AllowedTags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					gc.AllowedTags[t] = true
				}
			}
		}

		l.confs = append(l.confs, gc)
	}

	return l, outErr
}

// MustNew is like New but panics on error.
func MustNew(confs ...Config) *Logger {
	l, err := New(confs...)
	if err != nil {
		panic(err)
	}
	return l
}

// Close disables all destinations and releases their resources.
// The Logger must not be used after Close returns.
func (l *Logger) Close() {
	for _, gc := range l.confs {
		gc.Disabled = true
		gc.DestModule.Close()
		gc.DestModule = nil
	}
}

// Log writes a plain informational message.
// Optional variadic args may include an int (log level) and/or a string (tag).
func (l *Logger) Log(message interface{}, arg ...any) {
	l.doLog(message, false, false, arg...)
}

// Warn writes a warning message.
// Optional variadic args may include an int (log level) and/or a string (tag).
func (l *Logger) Warn(message interface{}, arg ...any) {
	l.doLog(message, false, true, arg...)
}

// Error writes an error message.
// Optional variadic args may include an int (log level) and/or a string (tag).
func (l *Logger) Error(message interface{}, arg ...any) {
	l.doLog(message, true, false, arg...)
}

// doLog is the common dispatch path for all log methods.
func (l *Logger) doLog(message interface{}, isErr bool, isWarn bool, args ...any) {
	if l.disabled {
		return
	}

	logLevel := 0
	logTag := ""

	// Extract optional level (int) and tag (string) from variadic args.
	for _, a := range args {
		switch v := a.(type) {
		case int:
			logLevel = v
		case string:
			logTag = v
		}
	}

	// Coerce the message to a string.
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	case []byte:
		msg = string(v)
	case int:
		msg = strconv.Itoa(v)
	case float64:
		msg = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			msg = "true"
		} else {
			msg = "false"
		}
	default:
		msg = ""
	}

	for _, gc := range l.confs {
		if gc.Disabled {
			continue
		}

		// Filter by message type.
		if gc.IgnoreLog && !isErr && !isWarn {
			continue
		}
		if gc.IgnoreError && isErr {
			continue
		}
		if gc.IgnoreWarn && isWarn {
			continue
		}

		// Filter by numeric level.
		if gc.IgnoreAbove != 0 && logLevel > gc.IgnoreAbove {
			continue
		}
		if gc.IgnoreBelow != 0 && logLevel < gc.IgnoreBelow {
			continue
		}

		// Filter by tag whitelist (only active when AllowedTags is non-empty).
		if len(gc.AllowedTags) != 0 {
			if logTag == "" {
				continue
			}
			if _, ok := gc.AllowedTags[logTag]; !ok {
				continue
			}
		}

		gc.DestModule.Write(core.FormatMessage(msg, gc), isErr, isWarn)
	}
}
