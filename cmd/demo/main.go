package main

import (
    "github.com/tominkoltd/go-logger"
	"time"
)

func main() {
	Console := logger.Logger{}
	Console.Log("test")


	time.Sleep(200 * time.Millisecond)
}