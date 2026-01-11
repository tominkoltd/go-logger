package file

import (
	"github.com/tominkoltd/go-logger/core"
	"os"
	"errors"
	"sync"
	"math"
)

type logMessage struct {
	message		string
}

type modData struct {
	disabled	bool
	fileName	string
	channel		chan logMessage
	config		*core.Config
	dropCounter	int
}

var (
	mtx			sync.Mutex
	channels 	= map[string]chan logMessage{}
	connections = map[string]int{}
)

func New(c *core.Config) (*modData, error) {
	md := &modData{}
	md.config = c

	if len(c.Destination) < 7 {
		md.disabled = true
		return md, errors.New("invalid file destination")		
	}

	fileName := c.Destination[5:]
	md.fileName = fileName

	mtx.Lock()
	defer mtx.Unlock()

	ch, ok := channels[fileName]
	if !ok {
		fh, err := os.OpenFile(
			fileName,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)

		if err != nil {
			md.disabled = true
			return md, err
		}

		ch = make(chan logMessage, c.QueueSize)
		channels[fileName] = ch
		go loggerWorker(fh, ch)
	}

	chc, cok := connections[fileName]
	if !cok {
		chc = 1
	} else {
		chc ++
	}
	connections[fileName] = chc

	md.channel = ch
	return md, nil
}

func (md *modData) Write(s string) error {
	if md.disabled {
		return errors.New("profile disabled")
	}
	if md.channel == nil {
		return errors.New("file channel not initialized")
	}
	if md.config.DropIfBusy {
		select {
			case md.channel <- logMessage{message: s}:
				return nil
			default: 	
				if md.dropCounter < math.MaxInt {
					md.dropCounter++
				}
				return nil
		}
	}
	md.channel <- logMessage{message: s}
	return nil
}

func (md *modData) Close() error {
	mtx.Lock()
	defer mtx.Unlock()

	cnt, ok := connections[md.fileName]
	if !ok || cnt <= 0 {
		return nil
	}

	cnt --
	if cnt == 0 {
		ch, ok := channels[md.fileName]
		if !ok  && ch != nil {
			close(ch)
		}
		delete(channels, md.fileName)
		delete(connections, md.fileName)
	} else {
		connections[md.fileName] = cnt
	}

	return nil
}


func loggerWorker(fileHandler *os.File, ch <-chan logMessage) {
	defer fileHandler.Close()
	for log := range ch {
		_, _ = fileHandler.WriteString(log.message)
		_, _ = fileHandler.WriteString("\n")
	}
}
