package parser

import (
	"testing"
	"time"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func BenchmarkTLI(b *testing.B) {
	log.SetLevel("debug")
	text := `
bpm 60

loop one
c6(h50) d e♭ f

loop two
g6 a b♭ c

loop chords
Cm(ru8d8,t30)
_
F

chain one
out crow(output=1,attack=0.1,decay=0.1,sustain=5,release=2)

chain chords

	`
	tli, _ := ParseText(text)
	log.SetLevel("info")
	maxDuration := 5 * time.Second
	startTime := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tli.Render()
		if time.Since(startTime) > maxDuration {
			break
		}
	}
}

func TestTLI(t *testing.T) {
	log.SetLevel("debug")
	text := `
loop one
Cm;1(ru4d4u2,h50,t120)
F;1(ru4d4u2,h50,t120)

loop two 
c1(h50) d e♭ f
f - a g

loop tuning 
c2

chain one
out crow(output=1,attack=0.1,decay=0.1,sustain=5,release=0.1)

chain two
out crow(output=3,attack=0.1,decay=0.1,sustain=5,release=0.1)

	`

	tli, err := ParseText(text)
	assert.Nil(t, err)
	err = tli.Render()
	assert.Nil(t, err)
	log.Debugf("tli: %+v", tli)

	done := make(chan bool)
	tli.Run(done)
	time.Sleep(82 * time.Second)
	done <- true
	time.Sleep(1 * time.Second)
}
