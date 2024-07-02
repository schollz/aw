package globals

import (
	"os"

	"github.com/schollz/aw/internal/parser"
	log "github.com/schollz/logger"
)

var TLI *parser.TLI

func ProcessFilename(filename string) (err error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		log.Error(err)
		return
	}
	text := string(b)
	TLI, err = parser.ParseText(text)
	if err != nil {
		log.Error(err)
		return
	}
	err = TLI.Render()
	if err != nil {
		log.Error(err)
		return
	}
	return
}
