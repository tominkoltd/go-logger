# go-logger

A lightweight, high-performance, configurable logger for Go with
multi-destination support, ANSI color parsing, and non-blocking output.

Designed to be simple, predictable, and fast — especially for CLI tools,
daemons, and process managers.

---

## Features

- Multiple logging destinations
- Per-destination configuration
- ANSI color parsing with custom tokens
- Optional ANSI stripping
- Line alignment for multi-line logs
- Non-blocking logging (optional drop-on-busy)
- Safe concurrent logging using goroutines + channels
- Clean shutdown with reference counting

---

## Supported Destinations (current)

✅ **stdout**  
✅ **stderr**  
✅ **file**

### Planned / TODO

- UDP
- TCP
- Unix socket (socket-file)
- Syslog (socket + UDP)

(Structure is in place; implementations are coming.)

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

	Console, err := logger.New(
		logger.Config{
			TimeStampFormat : timeFormat,
			Destination     : "file:test.log",
			ParseColors     : true,
			AlignLines      : true,
			IgnoreError     : true,
			IgnoreWarn      : true,
		},
		logger.Config{
			TimeStampFormat : timeFormat,
			ParseColors     : true,
			AlignLines      : true,
			IgnoreLog       : true,
			Prefix          : "ç88-[ç1-ERRORç88-]çR-",
		},
	)

	if err != nil {
		panic(err)
	}

	Console.Log("test", 0)
	Console.Log("test2ç21-coloredçR-", 1)
	Console.Log("test longer te\nfdgkjhd ç,5-dfkjgh dfkgjh dçR-fkjhg d\ndfkjgh2", 2)
	Console.Error("first errtest2 ERRRRR\nrdsftrgsdrftgw ertwer ter\nfgsdfg", 3)

	time.Sleep(200 * time.Millisecond)
	Console.Close()
}
```

---

## Destination Formats

```text
stdout
stderr
file:<file name with path>
udp:<ip/address:port>        (TODO)
tcp:<ip/address:port>        (TODO)
socket-file:<path>           (TODO)
syslog:<socket path>          (TODO)
syslog-udp:<host:port>        (TODO)
```

Examples:
```text
file:./debug.log
file:/var/log/myapp.log
```

---

## ANSI Color Tokens

The logger supports custom ANSI tokens using the `ParseColorChar`
(default: `ç`).

### Format
```text
ç<fg>,<bg>,<effect>-
```

Empty fields are ignored.

### Examples

```text
ç94-            → foreground only
ç94,5-          → foreground + background
ç94,,9-         → foreground + strikethrough
ç,,21-          → effect only (double underline)
çR-             → reset
```

Effects use **raw ANSI SGR codes**:

| Effect | Code |
|------|------|
| Bold | 1 |
| Dim | 2 |
| Italic | 3 |
| Underline | 4 |
| Blink | 5 |
| Reverse | 7 |
| Strikethrough | 9 |
| Overline | 53 |

---

## Configuration

```go
type Config struct {
	Destination       string
	TimeStampFormat   string
	QueueSize         int
	ReportInterval    int
	DropIfBusy        bool
	Prefix            string
	StripAnsi         bool
	AlignLines        bool
	AlignSeparator    rune
	AllowedTags       string
	IgnoreBelow       int
	IgnoreAbove       int
	ParseColors       bool
	ParseColorChar    rune
	IgnoreLog         bool
	IgnoreError       bool
	IgnoreWarn        bool
}
```

---

## Non-Blocking Mode

If `DropIfBusy` is enabled, log writes will **never block**.
Messages are dropped when the destination queue is full.

This is useful for:
- high-frequency logging
- process managers
- real-time systems

---

## Shutdown / Cleanup

Always call `Close()` when the logger is no longer needed:

```go
Console.Close()
```

This:
- flushes pending log messages
- closes file handlers
- stops worker goroutines cleanly

---

## Design Notes

- One goroutine per destination
- No locks in write paths
- Channels provide ordering and isolation
- Fail-open philosophy: logging never crashes your app

---

## License

MIT
