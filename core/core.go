// Package core contains the internal configuration type, the message-formatting
// pipeline, and the Destination interface implemented by each output protocol.
package core

import (
	"strings"
	"time"

	dateTime "github.com/tominkoltd/go-datetime"
	"github.com/tominkoltd/go-grapheme"
)

// Destination is the interface that every output protocol must implement.
type Destination interface {
	// Write delivers a single formatted log line to the destination.
	Write(string) error
	// Close flushes pending writes and releases any held resources.
	Close() error
}

// Config is the internal configuration record shared between the logger and its
// destination module.  It is populated by the public logger.Config and must not
// be modified after the destination has been initialised.
type Config struct {
	Destination     string
	TimeStampFormat string
	QueueSize       int
	ReportInterval  int
	DropIfBusy      bool
	Prefix          string
	StripAnsi       bool
	AlignLines      bool
	AlignSeparator  rune
	AllowedTags     map[string]bool
	IgnoreBelow     int
	IgnoreAbove     int
	ParseColors     bool
	ParseColorChar  string // trigger character for inline colour tokens
	DestModule      Destination
	IgnoreLog       bool
	IgnoreError     bool
	IgnoreWarn      bool
	Disabled        bool
}

// FormatMessage applies the full formatting pipeline to a raw message string
// and returns the final line ready for the destination writer.
//
// Pipeline order:
//  1. Prepend timestamp (if TimeStampFormat is set)
//  2. Prepend static prefix (if Prefix is set)
//  3. Expand colour tokens in the prefix (if ParseColors)
//  4. Indent continuation lines to align with the message body (if AlignLines)
//  5. Expand colour tokens in the message body (if ParseColors)
//  6. Strip ANSI escape sequences from the assembled line (if StripAnsi)
func FormatMessage(message string, gc *Config) string {
	prefix := ""

	if gc.TimeStampFormat != "" {
		prefix = dateTime.Format(time.Now(), gc.TimeStampFormat) + " "
	}
	if gc.Prefix != "" {
		prefix += gc.Prefix + " "
	}

	if gc.ParseColors {
		prefix = gc.parseColors(prefix)
	}

	if gc.AlignLines {
		// Count grapheme clusters in the rendered prefix so that multi-byte
		// characters (e.g. emoji) are measured correctly.
		tabSpace := grapheme.Count(prefix)
		var spacer string
		if gc.AlignSeparator != 0 && tabSpace > 2 {
			// Leave two columns for the separator character and a space.
			spacer = strings.Repeat(" ", tabSpace-2) + string(gc.AlignSeparator) + " "
		} else {
			spacer = strings.Repeat(" ", tabSpace)
		}
		message = strings.ReplaceAll(message, "\n", "\n"+spacer)
	}

	if gc.ParseColors {
		message = gc.parseColors(message)
	}

	ret := prefix + message

	if gc.StripAnsi {
		ret = gc.stripAnsi(ret)
	}

	return ret
}

// stripAnsi removes ANSI CSI colour/style escape sequences (ESC[...m) from s.
// Sequences that are malformed or unterminated are left in place and the
// original string is returned unchanged.
func (gc *Config) stripAnsi(in string) string {
	if !strings.Contains(in, "\033") {
		return in
	}

	lIndex := strings.IndexByte(in, 27) // position of first ESC
	var b strings.Builder
	pos := 0

	for {
		// Find the 'm' that terminates the current escape sequence.
		rIndex := strings.IndexByte(in[pos+lIndex:], 'm')
		if rIndex < 1 {
			// Malformed sequence — return the original string unmodified.
			return in
		}

		b.WriteString(in[pos : pos+lIndex]) // text before ESC
		pos = pos + lIndex + rIndex + 1     // advance past the 'm'

		lIndex = strings.IndexByte(in[pos:], 27)
		if lIndex == -1 {
			b.WriteString(in[pos:]) // remainder contains no more ESC
			break
		}
	}

	return b.String()
}

// parseColors expands inline colour tokens into ANSI 256-colour escape codes.
//
// Token syntax (where gc.ParseColorChar is the trigger, default 'ç'):
//
//	ç<n>-            → ESC[38;5;<n>m          (foreground colour)
//	ç<fg>,<bg>-      → ESC[38;5;<fg>;48;5;<bg>m
//	ç<fg>,<bg>,<a>-  → ESC[38;5;<fg>;48;5;<bg>;<a>m  (with attribute)
//	çR-              → ESC[0m                 (reset)
//
// The function operates in a single pass, handling reset and colour tokens
// inline without an intermediate string allocation.
func (gc *Config) parseColors(in string) string {
	needle := gc.ParseColorChar
	lIndex := strings.Index(in, needle)
	if lIndex == -1 {
		return in
	}

	needleLen := len(needle)
	var b strings.Builder
	b.Grow(len(in)) // pre-size to avoid incremental resizing
	pos := 0

	for {
		b.WriteString(in[pos : pos+lIndex]) // text before the trigger

		start := pos + lIndex + needleLen
		rIndex := strings.IndexByte(in[start:], '-')
		if rIndex < 0 {
			// No closing '-' — write the rest verbatim.
			b.WriteString(in[pos+lIndex:])
			break
		}

		token := in[start : start+rIndex]
		pos = start + rIndex + 1 // advance past the '-'

		switch {
		case token == "R":
			// Reset all attributes.
			b.WriteString("\033[0m")

		case !strings.Contains(token, ","):
			// Simple foreground-only colour.
			b.WriteString("\033[38;5;")
			b.WriteString(token)
			b.WriteByte('m')

		default:
			// Compound token: foreground, optional background, optional attribute.
			parts := strings.SplitN(token, ",", 3)
			b.WriteString("\033[")
			sep := ""
			if parts[0] != "" {
				b.WriteString("38;5;")
				b.WriteString(parts[0])
				sep = ";"
			}
			if len(parts) > 1 && parts[1] != "" {
				b.WriteString(sep)
				b.WriteString("48;5;")
				b.WriteString(parts[1])
				sep = ";"
			}
			if len(parts) > 2 && parts[2] != "" {
				b.WriteString(sep)
				b.WriteString(parts[2])
			}
			b.WriteByte('m')
		}

		lIndex = strings.Index(in[pos:], needle)
		if lIndex == -1 {
			b.WriteString(in[pos:]) // no more tokens
			break
		}
	}

	return b.String()
}
