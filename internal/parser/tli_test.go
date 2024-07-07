package parser

import (
	"testing"
	"time"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

// func BenchmarkTLI(b *testing.B) {
// 	log.SetLevel("debug")
// 	text := `
// bpm 60

// loop one
// c6(h50) d e♭ f

// loop two
// g6 a b♭ c

// loop chords
// Cm(ru8d8,t30)
// _
// F

// chain one
// out crow(output=1,attack=0.1,decay=0.1,sustain=5,release=2)

// chain chords

// 	`
// 	tli := New()
// 	tli.ParseText(text)
// 	log.SetLevel("info")
// 	maxDuration := 5 * time.Second
// 	startTime := time.Now()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		tli.Render()
// 		if time.Since(startTime) > maxDuration {
// 			break
// 		}
// 	}
// }

// func TestEmpty(t *testing.T) {
// 	log.SetLevel("debug")
// 	text := ``
// 	tli := New()
// 	err := tli.ParseText(text)
// 	assert.Nil(t, err)
// 	err = tli.Render()
// 	assert.Nil(t, err)
// 	log.Debugf("tli: %+v", tli)
// }

// func TestClock(t *testing.T) {

// 	log.SetLevel("debug")
// 	text := `
// loop clock
// c8(t180) c0 c8 c0 c8 c0 c8 c0 c8 c0 c8 c0 c8 c0 c8 c0

// chain clock
// out crow(4)
// `
// 	tli := New()
// 	err := tli.ParseText(text)
// 	assert.Nil(t, err)
// 	err = tli.Render()
// 	assert.Nil(t, err)

// 	tli.Play()
// 	time.Sleep(30 * time.Second)
// 	tli.Stop()

// }

// func TestArpy(t *testing.T) {
// 	log.SetLevel("debug")
// 	text := `
// loop test
// abea(h50,ru4d2,b32)

// loop chords
// Am;2(ru4d4u4d4)
// Am;2(ru4d4u4d4)
// Em;2(ru4d4u4d4)
// Em;2(ru4d4u4d4)
// Am;2(ru4d4u4d4)
// Am;2(ru4d4u4d4)
// F/A;2(ru4d4u4d4)
// F/A;2(ru4d4u4d4)

// chain test
// out crow(1)

// chain chords
// out crow(2)
// `
//
//		tli := New()
//		err := tli.ParseText(text)
//		assert.Nil(t, err)
//		err = tli.Render()
//		log.Debugf("tli: %+v", tli)
//		assert.Nil(t, err)
//		tli.Play()
//		time.Sleep(60 * time.Second)
//		tli.Stop()
//	}

func TestTLIArg(t *testing.T) {
	log.SetLevel("trace")
	text := `
loop a
a(adsr=(1,2)) b c

chain a
out crow(1)
`

	tli, err := New(text)
	assert.Nil(t, err)
	log.Debugf("tli: %+v", tli.Chains)
	log.Debugf("tli: %+v", tli.ChainsRendered)
	tli.Play()
	time.Sleep(3 * time.Second)
}

func TestTLIUpdate(t *testing.T) {
	log.SetLevel("debug")
	text := `

loop test
a(h50,t60,b8) b c d

chain test
out crow(output=1)
	`

	tli, err := New(text)
	assert.Nil(t, err)
	log.Debugf("tli: %+v", tli.Chains)
	log.Debugf("tli: %+v", tli.ChainsRendered)

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
