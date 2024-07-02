package crow

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	log "github.com/schollz/logger"
	"go.bug.st/serial"
)

var mutex sync.Mutex

type Crow struct {
	conn     serial.Port
	on       bool
	PortName string
}

type Murder struct {
	IsReady bool
	Crow    []Crow
}

func New() (m Murder, err error) {
	log.Trace("setting up crows")
	defer func() {
		log.Trace("crow setup")
	}()
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Error(err)
		return
	}
	crow := Crow{}
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	for _, port := range ports {
		crow.conn, err = serial.Open(port, mode)
		if err != nil {
			continue
		}
		_, err = crow.conn.Write([]byte("^^version"))
		if err != nil {
			log.Error(err)
			continue
		}
		crow.conn.SetReadTimeout(100 * time.Millisecond)
		// read response
		buf := make([]byte, 100)
		n, err := crow.conn.Read(buf)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Tracef("read %d bytes: %s", n, buf[:n])
		if bytes.Contains(buf[:n], []byte("v2")) || bytes.Contains(buf[:n], []byte("v3")) || bytes.Contains(buf[:n], []byte("v4")) {
			crow.PortName = port
			// setup default
			_, err = crow.conn.Write([]byte("^^First"))
			if err != nil {
				log.Error(err)
				continue
			}
			n, err = crow.conn.Read(buf)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Tracef("read %d bytes: %s", n, buf[:n])
			log.Debugf("crow connected on %s", crow.PortName)
			m.Crow = append(m.Crow, crow)
			crow = Crow{}
		}

	}
	log.Debugf("found %d crows", len(m.Crow))
	if len(m.Crow) > 0 {
		m.IsReady = true
	}
	return
}

func (m Murder) Close() (err error) {
	for _, crow := range m.Crow {
		errClose := crow.conn.Close()
		if errClose != nil {
			err = errClose
			log.Error(err)
		} else {
			log.Debugf("closed crow at %s", crow.PortName)
		}
	}
	return
}

// On switches the crow, 1-indexed
func (m *Murder) On(output int, on bool) (err error) {
	crowIndex := int(math.Floor(float64(output-1) / 4))
	if crowIndex >= len(m.Crow) {
		err = fmt.Errorf("output '%d' exceeds number of crows (%d)", output, len(m.Crow))
		return
	}
	output = ((output - 1) % 4) + 1
	cmd := fmt.Sprintf("output[%d](true)", output)
	if !on {
		cmd = fmt.Sprintf("output[%d](false)", output)

	}
	err = m.Command(crowIndex, cmd)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

type ADSR struct {
	Attack  float64
	Decay   float64
	Sustain float64
	Release float64
}

func (m *Murder) SetADSR(output int, adsr ADSR) (err error) {
	if !m.IsReady || len(m.Crow) < 1 {
		err = fmt.Errorf("not ready")
		return
	}
	crowIndex := int(math.Floor(float64(output-1) / 4))
	if crowIndex >= len(m.Crow) {
		err = fmt.Errorf("output '%d' exceeds number of crows (%d)", crowIndex, len(m.Crow))
		return
	}
	output = ((output - 1) % 4) + 1

	cmd := fmt.Sprintf("output[%d].action=adsr(%3.3f,%3.3f,%3.3f,%3.3f)", output, adsr.Attack, adsr.Decay, adsr.Sustain, adsr.Release)
	err = m.Command(crowIndex, cmd)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func (m Murder) Command(crowIndex int, cmd string) (err error) {
	if crowIndex >= len(m.Crow) {
		err = fmt.Errorf("crowIndex out of range: %d", crowIndex)
		return
	}

	log.Trace("[crow command] " + cmd)
	cmd = strings.TrimSpace(cmd) + "\n"
	_, err = m.Crow[crowIndex].conn.Write([]byte(cmd))
	if err != nil {
		log.Error(err)
		return
	}
	m.Crow[crowIndex].conn.SetReadTimeout(10 * time.Millisecond)
	buf := make([]byte, 100)
	n, err := m.Crow[crowIndex].conn.Read(buf)
	if err != nil {
		log.Error(err)
	} else if n > 0 {
		log.Tracef("read %d bytes: %s", n, buf[:n])
	}
	return
}

func (m *Murder) SetVoltage(output int, voltage float64) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	crowIndex := int(math.Floor(float64(output-1) / 4))
	if crowIndex >= len(m.Crow) {
		err = fmt.Errorf("output '%d' exceeds number of crows (%d)", output, len(m.Crow))
		return
	}
	output = ((output - 1) % 4) + 1
	log.Tracef("setting crow %d output %d to %3.2f volts", crowIndex, output, voltage)
	cmd := fmt.Sprintf("output[%d].volts=%2.3f\n", output, voltage)
	err = m.Command(crowIndex, cmd)
	return
}
