package core

import (
	"github.com/tominkoltd/go-datetime"
	"github.com/tominkoltd/go-grapheme"
	"time"
	"strings"
)

type Destination interface {
	Write(string) error
}

type Config struct {
	Destination		string
	TimeStampFormat	string
	QueueSize		int
	ReportInterval	int
	DropIfBusy		bool
	Prefix			string
	StripAnsi		bool
	AlignLines		bool
	AlignSeparator	rune
	AllowedTags		map[string]bool
	IgnoreBelow		int
	IgnoreAbove		int
	ParseColors		bool
	ParseColorChar	[]byte
	DestModule 		Destination
	IgnoreLog		bool
	IgnoreError		bool
	IgnoreWarn		bool
}

func FormatMessage(message string, gc* Config) string {
	prefix 		:= ""

	if gc.TimeStampFormat != "" {
		prefix = dateTime.Format(time.Now(), gc.TimeStampFormat) + " "
	}
	if gc.Prefix != "" {
		prefix += gc.Prefix + " "
	}

	if gc.AlignLines {
		tabSpace	:= grapheme.Count(prefix)
		spacer		:= ""
		if gc.AlignSeparator != 0 && tabSpace > 2 {
			tabSpace -= 2
			spacer = strings.Repeat(" ", tabSpace) + string(gc.AlignSeparator) + " "
		} else {
			spacer = strings.Repeat(" ", tabSpace)
		}
		message = strings.ReplaceAll(message, "\n", "\n" + spacer)
	}

	ret := prefix + message

	if gc.StripAnsi {
		// Strip Ansi --- todo

		// -------------------
	}

	return ret
}

func StripAnsi(in string) string {
	ret := ""

	return ret
}