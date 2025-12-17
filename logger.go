package logger

import (
	"github.com/tominkoltd/go-datetime"
	//"strings"
	"fmt"
	"time"
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
	message		interface{}
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
	if l.TimeStampFormat == "" {
		l.TimeStampFormat = "%Y-%d-%m"
	}

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

		out := dateTime.Format(time.Now(), "Year: %Y, Month: %M  %Y-%m-%d %H:%i:%s %r")
		fmt.Println(out)
		fmt.Println(log.message)
	}
}