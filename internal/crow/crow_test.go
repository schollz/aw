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
	err = m.SetADSR(2, ADSR{Attack: 0.1, Decay: 0.1, Sustain: 0.5, Release: 0.1})
	assert.Nil(t, err)
	for _, voltage := range []float64{10.0, 5.0, 0.0} {
		err = m.SetVoltage(1, voltage)
		m.On(2, true)
		m.Flush()
		assert.Nil(t, err)
		time.Sleep(3 * time.Second)
		m.On(2, false)
		m.Flush()
		time.Sleep(3 * time.Second)
	}
	assert.Nil(t, err)

	// shut down crow
	err = m.Close()
	assert.Nil(t, err)
}
