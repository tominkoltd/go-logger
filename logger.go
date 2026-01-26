/*
	v1.0.1
	Destination:
		stdout
		stderr
		file:<file name with path> 			(./debug.log)
		udp:<ip/address:port>				(udp.logger.com:987)
		tcp:<ip/address:port>				(tcp.logger.com:987)
		socket-file:<file name with path>	(./logger.sock)
		syslog:<socket path>				(/dev/log)
		syslog-udp:<host:port>				(localhost:514)
*/

package logger

import (
	"github.com/tominkoltd/go-logger/core"
	"github.com/tominkoltd/go-logger/protocols/stdout"
	"github.com/tominkoltd/go-logger/protocols/stderr"
	"github.com/tominkoltd/go-logger/protocols/file"
	"strings"
	"strconv"
	"errors"
)

type Logger struct {
	confs		[]*core.Config
	disabled	bool
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
	AllowedTags		string
	IgnoreBelow		int
	IgnoreAbove		int
	ParseColors		bool
	ParseColorChar	rune
	IgnoreLog		bool
	IgnoreError		bool
	IgnoreWarn		bool
}

func New(confs ...Config) (*Logger, error) {
	var outErr error
	l := &Logger{}

	if len(confs)==0 {
		return nil, errors.New("Config is missing")
	}

	for _, c := range confs {
		gc := &core.Config{}
		gc.Destination 		= c.Destination
		gc.TimeStampFormat 	= c.TimeStampFormat
		gc.QueueSize 		= c.QueueSize
		gc.ReportInterval 	= c.ReportInterval
		gc.DropIfBusy 		= c.DropIfBusy
		gc.Prefix 			= c.Prefix
		gc.StripAnsi 		= c.StripAnsi
		gc.AlignLines 		= c.AlignLines
		gc.AlignSeparator 	= c.AlignSeparator
		gc.IgnoreBelow 		= c.IgnoreBelow
		gc.IgnoreAbove 		= c.IgnoreAbove
		gc.ParseColors 		= c.ParseColors
		gc.IgnoreLog 		= c.IgnoreLog
		gc.IgnoreError 		= c.IgnoreError
		gc.IgnoreWarn 		= c.IgnoreWarn
		gc.AllowedTags 		= make(map[string]bool)
		
		if c.QueueSize == 0 {
			c.QueueSize = 100
		}
		if c.ParseColors {
			if c.ParseColorChar != 0 {
				gc.ParseColorChar = []byte(string(c.ParseColorChar))
			} else {
				gc.ParseColorChar = []byte("ç")
			}
		}
		if c.Destination == "stdout" || c.Destination == "" {
			modLink, err := stdout.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink
		} else if c.Destination == "stderr" {
			modLink, err := stderr.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink
		} else if strings.HasPrefix(c.Destination, "file:") {
			modLink, err := file.New(gc)
			if err != nil {
				outErr = err
				gc.Disabled = true
			}
			gc.DestModule = modLink
		}

		if c.AllowedTags != "" {
			for _, t := range strings.Split(c.AllowedTags, ",") {
				t = strings.TrimSpace(t)
				if t == "" {
					continue
				}
				gc.AllowedTags[t] = true
			}
		}

		l.confs = append(l.confs, gc)
	}
	
	return l, outErr
}

func MustNew(confs ...Config) (*Logger, error) {
	l, err := New(confs...)
	if err != nil {
		panic(err)
	}
	return l, nil
}

func (l *Logger) Close() {
	for _, gc := range l.confs {
		gc.Disabled = true
		gc.DestModule.Close()
		gc.DestModule=nil
	}
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
	if l.disabled {
		return
	}
	logLevel 	:= 0
	logTag 		:= ""

	msg			:= ""

	for _, a := range args {
		switch v := a.(type) {
		case int:
			logLevel = v
		case string:
			logTag = v
		}
	}

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
		if gc.IgnoreLog && !err && !warn {
			continue
		}
		if gc.IgnoreError && err {
			continue
		}
		if gc.IgnoreWarn && warn {
			continue
		}

		if gc.IgnoreAbove != 0 && logLevel > gc.IgnoreAbove {
			continue
		}
		if gc.IgnoreBelow != 0 && logLevel < gc.IgnoreBelow {
			continue
		}

		if logTag == "" {
			if len(gc.AllowedTags) != 0 {
				continue
			}
		} else {
			if _, ok := gc.AllowedTags[logTag]; !ok {
				continue
			}
		}

		msg = core.FormatMessage(msg, gc)

		gc.DestModule.Write(msg)

	}
}
