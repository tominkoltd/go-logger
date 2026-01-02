package stderr

import (
	"github.com/tominkoltd/go-logger/core"
	"os"
)

type logMessage struct {
	message		string
}

var logChannel = make(chan logMessage, 50)


type modData struct {
}

func init() {
	go loggerWorker()
}




func New(c *core.Config) (*modData, error) {
	md := &modData{}
	return md, nil
}

func (md *modData) Write(s string) error {
	logChannel <- logMessage{message: s}
	return nil
}





func loggerWorker() {
	for log := range logChannel {
		_, _ = os.Stderr.Write([]byte(log.message))
		_, _ = os.Stderr.Write([]byte{'\n'})
	}
}
