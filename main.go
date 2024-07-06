package main

import (
	"os"

	"github.com/schollz/aw/cmd/micro"
	"github.com/schollz/aw/internal/globals"
	"github.com/schollz/aw/internal/parser"
	log "github.com/schollz/logger"
)

func main() {
	// open file for writing
	log.SetLevel("trace")
	f, err := os.Create("micro.log")
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

	globals.TLI, err = parser.New(``)
	if err != nil {
		panic(err)
	}
	micro.Run()
}
