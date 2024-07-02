package crow

import (
	"testing"
	"time"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func TestCrow(t *testing.T) {
	log.SetLevel("trace")
	m, err := New()
	assert.Nil(t, err)
	if len(m.Crow) == 0 {
		return
	}
	for _, voltage := range []float64{10.0, 5.0, 0.0} {
		err = m.SetVoltage(1, voltage)
		assert.Nil(t, err)
		time.Sleep(1 * time.Second)
	}
	err = m.SetADSR(1, ADSR{Attack: 0.1, Decay: 0.1, Sustain: 0.5, Release: 0.1})
	assert.Nil(t, err)
	err = m.On(1, true)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	err = m.On(1, false)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)

	// shut down crow
	err = m.Close()
	assert.Nil(t, err)
}
