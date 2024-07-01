package main

import (
	"os"

	"github.com/schollz/aw/cmd/micro"
	log "github.com/schollz/logger"
)

func main() {
	// open file for writing
	log.SetLevel("debug")
	f, _ := os.Create("micro.log")
	log.SetOutput(f)

	micro.Run()
}
