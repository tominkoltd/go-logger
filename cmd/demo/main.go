package main

import (
    "github.com/tominkoltd/go-logger"
	"time"
	"fmt"
)

func main() {
	timeFormat := "ç214-%Y-%m-%d %Tç59-.%fçR-"
	Console, err := logger.New(
		logger.Config{
			TimeStampFormat		: timeFormat,
			ParseColors			: true,
			AlignLines			: true,
			IgnoreError			: true,
			IgnoreWarn			: true,
		},
		logger.Config{
			TimeStampFormat		: timeFormat,
			ParseColors			: true,
			AlignLines			: true,
			IgnoreLog			: true,
			Prefix				: "ç88-[ç1-ERRORç88-]çR-",
		},
	)

	if err != nil {
		fmt.Println(err)
	}
	
	Console.Log("Application starting", 0)

	Console.Log(
		"Loading configuration\n"+
			" - config file: ./config.yml\n"+
			" - environment: production",
		1,
	)

	Console.Log(
		"Connected services:\n"+
			" - ç34-databaseçR-\n"+
			" - ç33-cacheçR-\n"+
			" - ç36-message queueçR-",
		2,
	)

	Console.Log(
		"Processing request batch\n"+
			"request_id=ç92-8f3a21çR-\n"+
			"user_id=ç93-421çR-\n"+
			"items=12",
		3,
	)

	Console.Error(
		"Failed to process request\n"+
			"reason: ç31-timeout while waiting for database responseçR-\n"+
			"retrying in 500ms",
		4,
	)

	Console.Log("Shutdown complete", 5)	
	time.Sleep(200 * time.Millisecond)
}