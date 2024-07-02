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
	tli := New()
	tli.ParseText(text)
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

func TestEmpty(t *testing.T) {
	log.SetLevel("debug")
	text := ``
	tli := New()
	err := tli.ParseText(text)
	assert.Nil(t, err)
	err = tli.Render()
	assert.Nil(t, err)
	log.Debugf("tli: %+v", tli)
}

func TestTLI(t *testing.T) {
	log.SetLevel("debug")
	text := `

loop test
a(h50,t160,b8) b c d

chain test
out crow(output=1)

// loop one
// C;1(ru4d4u2,h50,t60)
// Am;1(ru4d4u2)
// Em;1(ru4d4u2)
// G/B;1(ru4d4u2)

// loop one2
// F;1(ru4d4u2,h50)
// F;1(ru4d4u2)
// Am;1(ru4d4u2)
// Dm;1(ru4d4u2)

// loop two 
// c4(h50,t60)
// g4 a4 - -
// c5 e5 c5 g5
// -

// loop three
// e3(t60) - - d3
// e3 - - d3
// e3 - - d3
// e3 - - d3


// loop three2
// f3(t60) a4 - -
// a2 g2 - -
// f3 - - -
// e3 - - -


// loop tuning 
// c2

// chain one * 2 one2 *2
// out crow(output=1,env=1,attack=0.1,decay=0.1,sustain=5,release=0.1)

// chain two
// out crow(output=3)

// chain three*2 three2*2
// out crow(output=4)

	`

	tli := New()
	err := tli.ParseText(text)
	assert.Nil(t, err)
	err = tli.Render()
	assert.Nil(t, err)

	tli.Play()
	time.Sleep(3 * time.Second)
	err = tli.Update(`
loop test
f(t180) e d c

chain test
out crow(output=1)`)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	tli.Stop()
	log.Debug(tli.Playing)
	time.Sleep(1 * time.Second)
}
