# go-logger

A lightweight, high-performance, configurable logger for Go with
multi-destination support, ANSI color parsing, and non-blocking output.

Designed to be simple, predictable, and fast — especially for CLI tools,
daemons, and process managers.

---

## Features

- Multiple logging destinations, each with independent configuration
- Per-destination level filtering (`IgnoreBelow` / `IgnoreAbove`)
- Per-destination type filtering (`IgnoreLog`, `IgnoreError`, `IgnoreWarn`)
- Tag-based whitelisting (`AllowedTags`)
- Inline ANSI color tokens with a configurable trigger character
- Optional ANSI stripping (`StripAnsi`)
- Line alignment for multi-line messages (`AlignLines`)
- Non-blocking mode — drop on busy instead of blocking the caller (`DropIfBusy`)
- Safe concurrent logging via goroutines + buffered channels
- Shared file handles with reference counting — multiple Logger instances can
  target the same file without conflicting

---

## Supported Destinations

| Destination | Status |
|---|---|
| `stdout` | ✅ |
| `stderr` | ✅ |
| `file:<path>` | ✅ |
| `udp:<host:port>` | planned |
| `tcp:<host:port>` | planned |
| `socket-file:<path>` | planned |
| `syslog:<socket>` | planned |
| `syslog-udp:<host:port>` | planned |

---

## Installation

```bash
go get github.com/tominkoltd/go-logger
```

---

## Basic Usage

```go
package main

import (
	"github.com/tominkoltd/go-logger"
	"time"
)

func main() {
	timeFormat := "ç214-%Y-%m-%d %Tç59-.%fçR-"

	// Two destinations: file (logs only) and stdout (errors/warnings only).
	console, err := logger.New(
		logger.Config{
			TimeStampFormat: timeFormat,
			Destination:     "file:test.log",
			ParseColors:     true,
			AlignLines:      true,
			IgnoreError:     true,
			IgnoreWarn:      true,
		},
		logger.Config{
			TimeStampFormat: timeFormat,
			ParseColors:     true,
			AlignLines:      true,
			IgnoreLog:       true,
			Prefix:          "ç88-[ç1-ERRORç88-]çR-",
		},
	)
	if err != nil {
		panic(err)
	}
	defer console.Close()

	console.Log("server started")
	console.Log("colored text: ç21-blueçR-", 1, "startup")
	console.Log("multi\nline\nmessage", 2)
	console.Error("something went wrong\nwith details here", 3)

	time.Sleep(200 * time.Millisecond)
}
```

### MustNew

Use `MustNew` when you want to panic on configuration errors rather than handle
them explicitly:

```go
console := logger.MustNew(logger.Config{
	Destination: "stdout",
	Prefix:      "[app]",
})
defer console.Close()
```

---

## Destination Formats

```
stdout
stderr
file:<path>           e.g.  file:./debug.log  or  file:/var/log/app.log
```

---

## Log Methods

```go
console.Log("message")              // informational
console.Warn("message")             // warning
console.Error("message")            // error
```

All three methods accept optional variadic arguments (in any order):

```go
console.Log("message", 2)           // with level
console.Log("message", "tag")       // with tag
console.Log("message", 2, "tag")    // level + tag
```

---

## Tag Filtering

Set `AllowedTags` to a comma-separated list to whitelist specific tags:

```go
logger.Config{
	AllowedTags: "http,db",
}
```

- When `AllowedTags` is **empty**, all messages are passed through regardless of tag.
- When `AllowedTags` is **set**, only messages carrying a matching tag are written;
  untagged messages are dropped.

---

## Level Filtering

```go
logger.Config{
	IgnoreBelow: 2,   // drop messages with level < 2
	IgnoreAbove: 5,   // drop messages with level > 5
}
```

Zero (the default) disables the corresponding bound.

---

## ANSI Color Tokens

When `ParseColors: true`, inline color tokens in both the `Prefix` and the
message body are expanded to ANSI 256-color escape codes.

### Token syntax

```
ç<fg>-                foreground color only        (256-color index)
ç<fg>,<bg>-           foreground + background
ç<fg>,<bg>,<attr>-    foreground + background + SGR attribute
çR-                   reset all attributes
```

Empty fields are skipped, so `ç,,1-` sets only the attribute (bold).

### Examples

```
ç94-            → foreground only
ç94,5-          → foreground + background
ç94,,9-         → foreground + strikethrough
ç,,21-          → effect only (double underline)
çR-             → reset
```

### SGR attribute reference

| Attribute | Code |
|---|---|
| Bold | 1 |
| Dim | 2 |
| Italic | 3 |
| Underline | 4 |
| Blink | 5 |
| Reverse | 7 |
| Strikethrough | 9 |
| Overline | 53 |

### Custom trigger character

The default trigger is `ç`. Override it with `ParseColorChar`:

```go
logger.Config{
	ParseColors:    true,
	ParseColorChar: '§',
}
```

---

## Configuration Reference

```go
type Config struct {
	// Output destination: "stdout" (default), "stderr", "file:<path>"
	Destination string

	// Go time-layout string prepended to every message. Empty = no timestamp.
	TimeStampFormat string

	// Channel buffer depth (default 100).
	QueueSize int

	// Reserved for future use.
	ReportInterval int

	// Drop messages silently when the queue is full instead of blocking.
	DropIfBusy bool

	// Static string inserted after the timestamp and before the message.
	Prefix string

	// Strip ANSI escape sequences from the final output line.
	StripAnsi bool

	// Indent continuation lines of multi-line messages to align with the
	// message body (past the prefix).
	AlignLines bool

	// Character inserted between the prefix column and the continuation
	// indent, e.g. '|'. Zero = plain spaces.
	AlignSeparator rune

	// Comma-separated tag whitelist. Empty = pass everything.
	AllowedTags string

	// Drop messages with level strictly below this value. Zero = no lower bound.
	IgnoreBelow int

	// Drop messages with level strictly above this value. Zero = no upper bound.
	IgnoreAbove int

	// Expand inline color tokens to ANSI escape codes.
	ParseColors bool

	// Trigger character for color tokens (default 'ç').
	ParseColorChar rune

	// Suppress plain Log() messages on this destination.
	IgnoreLog bool

	// Suppress Error() messages on this destination.
	IgnoreError bool

	// Suppress Warn() messages on this destination.
	IgnoreWarn bool
}
```

---

## Non-Blocking Mode

With `DropIfBusy: true`, log writes **never block** the caller.
Messages are silently discarded when the destination queue is full.

Useful for high-frequency logging, process managers, and real-time systems
where a slow disk or network must not stall the application.

---

## Shutdown / Cleanup

Always call `Close()` when done — or use `defer`:

```go
defer console.Close()
```

`Close()` disables all destinations, drains and closes their channels, and
stops the worker goroutines. File handles are released once the last consumer
for a given path closes.

---

## Design Notes

- One worker goroutine per destination
- No locks in write paths — channels provide ordering and isolation
- File destinations use a shared handle + reference count so multiple Logger
  instances targeting the same path share one goroutine and one file descriptor
- Fail-open: logging failures never panic the application

---

## Changelog

### v1.0.3
- **Fix** `file.Close()`: channel was never closed due to an inverted condition
  (`!ok && ch != nil` → `ok && ch != nil`), causing a goroutine leak on every close
- **Fix** `parseColors`: panic when all needle occurrences were reset tokens
  (`lIndex == -1` after `ReplaceAll` caused an `in[-1:]` slice)
- **Fix** `New`: `QueueSize` default (100) was applied after copying to the internal
  config, leaving file destinations with an unbuffered channel
- **Fix** tag filtering: tagged messages were dropped even when `AllowedTags` was
  empty — filtering is now a no-op when no tags are configured
- **Fix** `file.New`: destination length check was off-by-one, rejecting single-char paths
- **Perf** `parseColors` rewritten as a single-pass loop — eliminates the
  `ReplaceAll` pre-scan, intermediate string allocation, and per-token concatenations
- **Perf** `ParseColorChar` changed from `[]byte` to `string` — removes a per-call conversion
- `MustNew` now returns `*Logger` (panicking constructors don't return an error)
- Full godoc coverage on all exported and unexported symbols

### v1.0.2
- Initial public release

---

## License

MIT
