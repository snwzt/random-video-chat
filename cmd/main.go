package main

import (
	"os"
	"snwzt/random-video-chat/cmd/sugar"
)

func main() {
	sugar.Execute(
		os.Exit,
		os.Args[1:],
	)
}
