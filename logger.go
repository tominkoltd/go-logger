package logger

import (
	"github.com/tominkoltd/go-datetime"
	"strings"
	"fmt"
	"time"
	"strconv"
)

const (
	EffectNone				= 0
	EffectBold       		= 1 << iota
	EffectDim
	EffectItalic
	EffectUnderline
	EffectBlink
	EffectReverse
	EffectStrikethrough
	EffectOverline
	EffectDoubleUnderline
)

type Logger struct {
	FileName		string
	TimeStampFormat	string
	DropIfBusy		bool
	DefaultLevel	int
	initialised		bool
}

type logMessage struct {
	l 			*Logger
	message		string
	isError 	bool
	isWarning	bool
	level		int
	prefix		string
}

var logChannel = make(chan logMessage, 50)

func init() {
	go loggerWorker()
}

func (l *Logger) Init() {
	l.initialised = true
}

func (l *Logger) Log(message interface{}, arg ...any) {
	l.doLog(message, false, false, arg...)
}
func (l *Logger) Warn(message interface{}, arg ...any) {
	l.doLog(message, false, true, arg...)
}
func (l *Logger) Error(message interface{}, arg ...any) {
	l.doLog(message, true, false, arg...)
}

func (l *Logger) doLog(message interface{}, err bool, warn bool, args ...any) {
	if !l.initialised {
		l.Init()
	}

	logLevel 	:= l.DefaultLevel
	logPrefix 	:= ""

	msg			:= ""

	for _, a := range args {
		switch v := a.(type) {
		case int:
			logLevel = v
		case string:
			logPrefix = v
		}
	}

	switch v := message.(type) {
        case string:
            msg = v
		case []byte:
			msg = string(v)		
        case int, float64, bool:
            msg = fmt.Sprintf("%v", v)
        default:
            msg = ""

	}


	logChannel <- logMessage{l: l, message: msg, isError: err, isWarning: warn, level: logLevel, prefix: logPrefix}
}


func loggerWorker() {
	for log := range logChannel {
		ts := ""
		if log.l.TimeStampFormat != "" {
			ts = dateTime.Format(time.Now(), log.l.TimeStampFormat)
		}
		fmt.Println(ts+" "+log.message)
	}
}

func GetAnsi(color int, background int, effect int) string {
	if color == 0 && background == 0 && effect == 0 {
		return "\x1b[0m"
	}
	var code strings.Builder
	code.WriteString("\033[")

	if effect == 0 {
		code.WriteString("0;")
	}
	if effect&EffectBold != 0 {
		code.WriteString("1;")
	}
	if effect&EffectDim != 0 {
		code.WriteString("2;")
	}
	if effect&EffectItalic != 0 {
		code.WriteString("3;")
	}
	if effect&EffectUnderline != 0 {
		code.WriteString("4;")
	}
	if effect&EffectBlink != 0 {
		code.WriteString("5;")
	}
	if effect&EffectReverse != 0 {
		code.WriteString("7;")
	}
	if effect&EffectStrikethrough != 0 {
		code.WriteString("9;")
	}
	if effect&EffectOverline != 0 {
		code.WriteString("53;")
	}
	if effect&EffectDoubleUnderline != 0 {
		code.WriteString("21;")
	}
	if color > 0 {
		code.WriteString("38;5;" + strconv.Itoa(color) + ";")
	} else {
		code.WriteString("39;")
	}
	if background > 0 {
		code.WriteString("48;5;" + strconv.Itoa(background) + ";")
	} else {
		code.WriteString("49;")
	}
	if code.Len() == 0 {
		return ""
	}
	b := []byte(code.String())
	b[len(b)-1] = 'm'
	return string(b)
}