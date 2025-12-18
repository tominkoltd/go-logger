package main

import (
    "github.com/tominkoltd/go-logger"
	"time"
)

func main() {
	format := 	logger.GetAnsi(100, 0, 0) + 
				"%Y-%m-%d %T." + 
				logger.GetAnsi(125, 0, 0) + 
				"%f" + 
				logger.GetAnsi(0, 0, 0)

	Console := logger.Logger{TimeStampFormat: format}
	
	Console.Log("test")
	Console.Log("test2")
	Console.Log("test2")
	Console.Log("test2")


	time.Sleep(200 * time.Millisecond)
}