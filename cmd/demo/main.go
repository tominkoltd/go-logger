package main

import (
    "github.com/tominkoltd/go-logger"
	"time"
	"fmt"
)

func main() {
	Console, err := logger.New(
		logger.Config{
			TimeStampFormat		: "%Y-%m-%d %T.%f",
			ParseColors			: true,
			ParseColorChar		: 'ç',
			AlignLines			: true,
			AlignSeparator		: '|',
			IgnoreError			: true,
			IgnoreWarn			: true,
		},
		logger.Config{
			TimeStampFormat		: "%Y-%m-%d %T.%f",
			ParseColors			: true,
			ParseColorChar		: 'ç',
			AlignLines			: true,
			AlignSeparator		: '|',
			IgnoreLog			: true,
			Prefix				: "[ERROR]",
		},
	)

	if err != nil {
		fmt.Println(err)
	}
	
	Console.Log("test", 0)
	Console.Log("test2ç21-coloredçR-", 1)
	Console.Log("test longer te\nfdgkjhd dfkjgh dfkgjh dfkjhg d\ndfkjgh2", 2)
	Console.Error("test2 ERRRRR\nrdsftrgsdrftgw ertwer ter\nfgsdfg", 3)
	Console.Log("test5", 4)
	Console.Log("test6", 5)
	Console.Log("test longer te\nfdgkjhd dfkjgh dfkgjh dfkjhg d\ndfkjgh2", 6)

	time.Sleep(200 * time.Millisecond)
}