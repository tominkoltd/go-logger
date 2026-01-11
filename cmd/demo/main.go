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
			Destination			: "file:test.log",
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
	
	Console.Log("test", 0)
	Console.Log("test2ç21-coloredçR-", 1)
	Console.Log("test longer te\nfdgkjhd ç,5-dfkjgh dfkgjh dçR-fkjhg d\ndfkjgh2", 2)
	Console.Error("first errtest2 ERRRRR\nrdsftrgsdrftgw ertwer ter\nfgsdfg", 3)

	//Console.Close()

	Console.Error("test2 ERRRRR\nrdsftrgsdrftgw ertwer ter\nfgsdfg", 3)
	Console.Log("test5", 4)
	Console.Log("test6", 5)
	Console.Log("test longer te\nfdgkjhd dfkjgh dfkgjh dfkjhg d\ndfkjgh2", 6)
	
	time.Sleep(200 * time.Millisecond)
}