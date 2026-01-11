package core

import (
	"github.com/tominkoltd/go-datetime"
	"github.com/tominkoltd/go-grapheme"
	"time"
	"strings"
)

type Destination interface {
	Write(string) error
	Close() error
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
	Disabled		bool
}

func FormatMessage(message string, gc* Config) string {
	prefix 		:= ""

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

	if gc.ParseColors {
		message = gc.parseColors(message)
	}
	ret := prefix + message

	if gc.StripAnsi {
		ret = gc.stripAnsi(ret)
	}

	return ret
}

func (gc* Config) stripAnsi(in string) string {
	if !strings.Contains(in, "\033") {
		return in
	}

	lIndex := strings.IndexByte(in, 27)
	var b strings.Builder	

	pos := 0
	for {
		rIndex := strings.IndexByte(in[pos+lIndex:], 'm')
		if rIndex < 1 {
			return in
		}
		b.WriteString(in[pos:pos+lIndex])
		pos = pos+lIndex+rIndex+1
		lIndex = strings.IndexByte(in[pos:], 27)
		if lIndex == -1 {
			b.WriteString(in[pos:])
			break
		}
	}
	return b.String()
}

func (gc* Config) parseColors(in string) string {
	needle := string(gc.ParseColorChar)
	if !strings.Contains(in, needle) {
		return in
	}
	in = strings.ReplaceAll(in, needle + "R-", "\033[0m")
	lIndex := strings.Index(in, needle)

	var b strings.Builder
	
	pos := 0
	for {
		rIndex := strings.IndexByte(in[pos+lIndex:], '-')
		if rIndex < 1 {
			return in
		}
		token := in[pos+lIndex+len(needle):pos+lIndex+rIndex]
		
		b.WriteString(in[pos:pos+lIndex])
		pos = pos+lIndex+rIndex+1

		if !strings.Contains(token, ",") {
			b.WriteString("\033[38;5;" + token + "m")
		} else {
			color := "\033["
			colorTokents := strings.Split(token, ",")
			if (colorTokents[0] != "") {
				color += "38;5;" + colorTokents[0] + ";"
			}
			if len(colorTokents) > 1 && colorTokents[1] != "" {
				color += "48;5;" + colorTokents[1] + ";"
			}
			if len(colorTokents) > 2 && colorTokents[2] != ""{
				color += colorTokents[2] + ";"
			}
			if color[len(color)-1] == ';' {
				color = color[:len(color)-1]
			}
			b.WriteString(color +"m")
		}
		lIndex = strings.Index(in[pos:], needle)
		if lIndex == -1 {
			b.WriteString(in[pos:])
			break
		}
	}
	
	return b.String()
}