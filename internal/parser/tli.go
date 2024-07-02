package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/loov/hrtime"
	"github.com/schollz/aw/internal/crow"
	log "github.com/schollz/logger"
)

var crows crow.Murder
var mutex sync.Mutex

type TLI struct {
	Chains  []Chain `json:"chains"`
	Loops   []Loop  `json:"loops"`
	Params  Params  `json:"params"`
	Playing bool    `json:"playing"`
}

type Params struct {
	IsSet    int     `json:"isset"`
	Tempo    int     `json:"tempo"`
	Gate     float64 `json:"gate"`
	Velocity float64 `json:"velocity"`
}

const (
	TempoSet = 1 << iota
	GateSet
	VelocitySet
)

func (p *Params) Set(param int, val interface{}) {
	p.IsSet |= param
	switch param {
	case TempoSet:
		p.Tempo = val.(int)
	case GateSet:
		p.Gate = val.(float64)
	case VelocitySet:
		p.Velocity = val.(float64)
	}
}

func (p Params) CheckSet(param int) bool {
	return p.IsSet&param != 0
}

func (t TLI) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("params: %+v", t.Params))
	for _, c := range t.Chains {
		sb.WriteString("\n")
		sb.WriteString(c.String())
	}
	for _, p := range t.Loops {
		sb.WriteString(fmt.Sprintf("\nloop '%s':", p.Name))
		for _, s := range p.Steps {
			sb.WriteString("\n")
			sb.WriteString(s.String())
		}
	}
	return sb.String()
}

type Chain struct {
	NameLoop          []string   `json:"loops"`
	Outs              []string   `json:"outs"`
	OutFns            []Function `json:"out_fns"`
	Steps             []Step     `json:"steps"` // filled in with Render()
	BeatsTotal        float64    `json:"beats_total"`
	MicrosecondsTotal int64      `json:"microseconds_total"`
	TimePosition      int64      `json:"time_position,omitempty"`
}

func (c Chain) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("chain: %v", c.NameLoop))
	sb.WriteString(fmt.Sprintf("\nbeats total: %v", c.BeatsTotal))
	sb.WriteString(fmt.Sprintf("\nmicroseconds total: %v", c.MicrosecondsTotal))
	for _, s := range c.Steps {
		sb.WriteString("\n")
		sb.WriteString(s.String())
	}

	return sb.String()
}

func (c Chain) PlayNote(notes []Note, on bool) (err error) {
	for _, out := range c.OutFns {
		log.Debugf("[%+v] [%+v] note %v: %+v", c.NameLoop, out, on, notes)
		switch out.Name {
		case "crow":
			// crow out
			var output int
			output, err = out.GetIntPlace("output", 0)
			if err != nil {
				log.Error(err)
				return
			}
			if crows.IsReady {
				for i, note := range notes {
					j := i * 2
					if on {
						crows.SetVoltage(output+j, float64(note.Midi)/12.0)
					}
					if crows.UseEnv[output+j] > 0 {
						crows.On(output+j+1, on)
					}
				}
			}
		case "sc":
			// sc out
		}
	}

	// for _, note := range notes {
	// 	sc.Send(sc.Options{Instrument: "jp2", Note: note.Midi, On: true})
	// }
	return
}

type Loop struct {
	Name             string `json:"name"`
	Steps            []Step `json:"steps"`
	lastMidiNote     int
	lastBeatsPerLine int
}

func LoopNew() Loop {
	return Loop{Name: "default", lastMidiNote: 60, lastBeatsPerLine: 4}
}

type Step struct {
	BeatsStart               float64 `json:"beats_start"`
	BeatsDuration            float64 `json:"beats_duration,omitempty"`
	BeatsPerLine             int     `json:"beats_per_line,omitempty"`
	StepLineCount            int     `json:"duration_proportion,omitempty"`
	TimeStartMicroseconds    int64   `json:"time_start"`
	TimeDurationMicroseconds int64   `json:"time_duration,omitempty"`
	Notes                    []Note  `json:"notes,omitempty"`
	IsNote                   bool    `json:"is_note,omitempty"`
	Token                    string  `json:"token,omitempty"`
	Arguments                []Arg   `json:"arguments,omitempty"`
	Params                   Params  `json:"params"`
}

func (s Step) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type Note struct {
	Midi         int    `json:"midi,omitempty"`
	Name         string `json:"name,omitempty"`
	NameOriginal string `json:"name_original,omitempty"`
	IsRest       bool   `json:"is_rest,omitempty"`
	IsLegato     bool   `json:"is_legato,omitempty"`
}

func NoteAdd(note Note, interval int) (result Note) {
	result = Note{Midi: note.Midi + interval, Name: note.Name}
	for _, d := range noteDB {
		if d.MidiValue == note.Midi+interval {
			result = Note{Midi: d.MidiValue, Name: d.NameSharp}
			break
		}
	}
	return
}

const (
	StateNone = iota
	StateLoop
	StateChain
	StateSet
)

func New() *TLI {
	tli := new(TLI)
	tli.Params = Params{Tempo: 120}
	return tli
}

func (tli *TLI) Update(text string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	tliTest := New()
	err = tliTest.ParseText(text)
	if err != nil {
		log.Error(err)
		return

	}
	err = tliTest.Render()
	if err != nil {
		log.Error(err)
		return
	}
	tli.ParseText(text)
	tli.Render()
	return
}

func (tli *TLI) ParseText(text string) (err error) {
	lines := strings.Split(text, "\n")
	loop := LoopNew()
	chain := Chain{}
	tli.Loops = []Loop{}
	tli.Chains = []Chain{}
	fnFinish := func() {
		if len(loop.Steps) > 0 {
			tli.Loops = append(tli.Loops, loop)
			loop = LoopNew()
		}
		if len(chain.NameLoop) > 0 {
			tli.Chains = append(tli.Chains, chain)
			chain = Chain{}
		}
	}
	// look for loop
	state := StateNone
	for _, line := range lines {
		// skip comments
		line = strings.Split(line, "//")[0]
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "loop") {
			fnFinish()
			state = StateLoop
			loop.Name = strings.Split(line, " ")[1]
			continue
		} else if strings.HasPrefix(line, "chain") {
			fnFinish()
			state = StateChain
			chain.NameLoop, err = ParseChain(line)
			log.Debugf("parsed chain: '%s' -> %+v", line, chain.NameLoop)
			if err != nil {
				log.Error(err)
				return
			}
			// parse chain
		} else if strings.HasPrefix(line, "set") {
			state = StateSet
		} else {
			switch state {
			case StateLoop:
				err = loop.AddLine(line)
				if err != nil {
					log.Error(err)
				}
			case StateChain:
				// parse chain
				if strings.HasPrefix(line, "out") {
					chain.Outs = strings.Fields(line)[1:]
				}
			case StateSet:
				// parse set
			}
		}
	}
	fnFinish()
	if len(loop.Steps) > 0 {
		tli.Loops = append(tli.Loops, loop)
	}
	if len(tli.Chains) == 0 {
		// make a chain of all current loops
		chain := Chain{}
		for _, loop := range tli.Loops {
			chain.NameLoop = append(chain.NameLoop, loop.Name)
		}
		tli.Chains = append(tli.Chains, chain)
	}

	return

}

func (p *Loop) AddLine(line string) (err error) {
	line = ExpandMultiplication(line)
	tokens, err := TokenizeLineString(line)
	if err != nil {
		log.Error(err)
		return
	}
	tokens, err = RetokenizeArpeggioArgument(tokens)
	if err != nil {
		log.Error(err)
		return
	}
	tokens, err = TokenizeLineString(TokenExpandToLine(tokens))
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("tokens: %+v", tokens)
	steps := []Step{}
	for _, token := range tokens {
		fn, _ := ParseFunction(token)
		log.Debugf("fn: %v, args: %v", fn.Name, fn.Args)
		notes, errPhrase := ParseChord(fn.Name, p.lastMidiNote)
		if errPhrase != nil {
			notes, errPhrase = ParseMidi(fn.Name, p.lastMidiNote)
		}
		step := Step{BeatsPerLine: p.lastBeatsPerLine, Token: token}
		step.Arguments = fn.Args
		if errPhrase == nil {
			log.Debugf("notes: %+v", notes)
			p.lastMidiNote = notes[len(notes)-1].Midi
			step.Notes = notes
		} else {
			// check for rest or legato
			if fn.Name == "_" {
				step.Notes = []Note{{IsLegato: true}}
			} else if fn.Name == "~" {
				step.Notes = []Note{{IsRest: true}}
			} else {
				continue
			}
		}
		for i := 0; i < len(fn.Args); i++ {
			decorator := fn.Args[i].Value
			if strings.HasPrefix(decorator, "t") {
				tempo, errParse := strconv.Atoi(decorator[1:])
				if errParse == nil {
					step.Params.Set(TempoSet, tempo)
				}
			} else if strings.HasPrefix(decorator, "b") {
				beats, errParse := strconv.Atoi(decorator[1:])
				if errParse == nil {
					step.BeatsPerLine = beats
					p.lastBeatsPerLine = beats
				}
			} else if strings.HasPrefix(decorator, "v") {
				velocity, errParse := strconv.ParseFloat(decorator[1:], 64)
				if errParse == nil {
					step.Params.Set(VelocitySet, velocity)
				}
			} else if strings.HasPrefix(decorator, "h") {
				gate, errParse := strconv.ParseFloat(decorator[1:], 64)
				if errParse == nil {
					step.Params.Set(GateSet, gate/100.0)
				}
			}
		}
		step.StepLineCount = len(tokens)
		steps = append(steps, step)
	}
	p.Steps = append(p.Steps, steps...)
	return
}

func (tli *TLI) Render() (err error) {

	for i := 0; i < len(tli.Chains); i++ {
		for _, loopName := range tli.Chains[i].NameLoop {
			for _, loop := range tli.Loops {
				if loop.Name == loopName {
					tli.Chains[i].Steps = append(tli.Chains[i].Steps, loop.Steps...)
				}
			}
		}
		// check if there are any steps
		if len(tli.Chains[i].Steps) == 0 {
			continue
		}
		// set the tempo on each step
		lastTempo := tli.Params.Tempo
		for j := 0; j < len(tli.Chains[i].Steps); j++ {
			if tli.Chains[i].Steps[j].Params.CheckSet(TempoSet) {
				log.Tracef("setting tempo to %d from step: %+v", tli.Chains[i].Steps[j].Params.Tempo, tli.Chains[i].Steps[j])
				lastTempo = tli.Chains[i].Steps[j].Params.Tempo
			}
			tli.Chains[i].Steps[j].Params.Tempo = lastTempo
		}
		// set the gate on each step
		lastGate := 0.95
		for j := 0; j < len(tli.Chains[i].Steps); j++ {
			if tli.Chains[i].Steps[j].Params.CheckSet(GateSet) {
				lastGate = tli.Chains[i].Steps[j].Params.Gate
			}
			tli.Chains[i].Steps[j].Params.Gate = lastGate
		}
		tli.Chains[i].Render()
	}
	// setup outputs
	for _, chain := range tli.Chains {
		for _, fn := range chain.OutFns {
			switch fn.Name {
			case "crow":
				if !crows.IsReady {
					crows, err = crow.New()
					if err != nil {
						log.Error(err)
					}
				}
				if crows.IsReady {
					var val float64
					var err error
					var adsr crow.ADSR
					var output int
					output, _ = fn.GetIntPlace("output", 0)
					if output > 0 {
						crows.UseEnv[output-1], _ = fn.GetInt("env")

						adsr = crow.ADSR{Attack: 0.1, Decay: 0.1, Sustain: 5, Release: 1}
						if val, err = fn.GetFloat("attack"); err == nil {
							adsr.Attack = val
						}
						if val, err = fn.GetFloat("decay"); err == nil {
							adsr.Decay = val
						}
						if val, err = fn.GetFloat("sustain"); err == nil {
							adsr.Sustain = val
						}
						if val, err = fn.GetFloat("release"); err == nil {
							adsr.Release = val
						}
						crows.SetADSR(output+1, adsr)
					}

				}
			}
		}
	}

	return
}

func (c *Chain) Render() {
	// figure out the beats alloted to each
	beatsTotal := 0.0
	microSecondsTotal := int64(0)
	for i := 0; i < len(c.Steps); i++ {
		for _, note := range c.Steps[i].Notes {
			if !note.IsRest && !note.IsLegato {
				c.Steps[i].IsNote = true
				break
			}
		}
		if c.Steps[i].IsNote {
			c.Steps[i].BeatsStart = beatsTotal
			c.Steps[i].TimeStartMicroseconds = microSecondsTotal
			c.Steps[i].BeatsDuration = float64(c.Steps[i].BeatsPerLine) / float64(c.Steps[i].StepLineCount)
			c.Steps[i].TimeDurationMicroseconds = int64(c.Steps[i].BeatsDuration * float64(60000000) / float64(c.Steps[i].Params.Tempo))
			// find how many steps until the next not or rest
			isStop := false
			for jj := i + 1; jj < len(c.Steps)*2; jj++ {
				j := jj
				for {
					if j < len(c.Steps) {
						break
					}
					j -= len(c.Steps)
				}
				for _, note := range c.Steps[j].Notes {
					if (!note.IsRest && !note.IsLegato) || note.IsRest {
						isStop = true
						break
					}
				}
				if isStop {
					break
				}
				c.Steps[i].BeatsDuration += float64(c.Steps[j].BeatsPerLine) / float64(c.Steps[j].StepLineCount)
				c.Steps[i].TimeDurationMicroseconds += int64(float64(c.Steps[j].BeatsPerLine) / float64(c.Steps[j].StepLineCount) * float64(60000000) / float64(c.Steps[j].Params.Tempo))
			}
		}
		beatsTotal += float64(c.Steps[i].BeatsPerLine) / float64(c.Steps[i].StepLineCount)
		microSecondsTotal += int64(float64(c.Steps[i].BeatsPerLine) / float64(c.Steps[i].StepLineCount) * float64(60000000) / float64(c.Steps[i].Params.Tempo))
	}
	// remove all steps that don't have beats_start
	newSteps := []Step{}
	for _, step := range c.Steps {
		if step.IsNote {
			newSteps = append(newSteps, step)
		}
	}
	// determine all the out functions
	for _, out := range c.Outs {
		fn, errFn := ParseFunction(out)
		if errFn != nil {
			log.Error(errFn)
			continue
		}
		c.OutFns = append(c.OutFns, fn)
	}

	c.Steps = newSteps
	c.BeatsTotal = beatsTotal
	c.MicrosecondsTotal = microSecondsTotal
}

func (tli *TLI) Toggle() {
	if tli.Playing {
		tli.Stop()
	} else {
		tli.Play()
	}
}
func (tli *TLI) Stop() {
	if tli.Playing {
		log.Debugf("stopping")
		tli.Playing = false
	}
}

func (tli *TLI) Play() {
	if len(tli.Chains) > 0 {
		go tli.run()
	}
}

// create a Run that takes an incoming bool channel to stop
func (tli *TLI) run() {
	if tli.Playing {
		return
	}
	tli.Playing = true
	ticker := time.NewTicker(10 * time.Microsecond)
	startTime := hrtime.Now()
	for i := range tli.Chains {
		tli.Chains[i].TimePosition = -1
	}
	go func() {
		// catch panic
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()

		for {
			select {
			case <-ticker.C:
				if !tli.Playing {
					log.Debug("not playing")
					return
				}
				mutex.Lock()
				for i, chain := range tli.Chains {
					// skip if no steps
					if len(chain.Steps) == 0 {
						continue
					}
					timePosition := hrtime.Since(startTime).Microseconds()
					for {
						if timePosition < chain.MicrosecondsTotal {
							break
						}
						timePosition -= chain.MicrosecondsTotal
					}
					for stepi, step := range chain.Steps {
						if !tli.Playing {
							mutex.Unlock()
							return
						}
						if (timePosition > step.TimeStartMicroseconds && tli.Chains[i].TimePosition <= step.TimeStartMicroseconds) ||
							(timePosition < tli.Chains[i].TimePosition && stepi == 0) {
							tli.Chains[i].PlayNote(step.Notes, true)
							go func(s Step) {
								sleepMS := int64(math.Round(float64(s.TimeDurationMicroseconds) * s.Params.Gate))
								sleepStart := hrtime.Now()
								for {
									if hrtime.Since(sleepStart).Microseconds() > sleepMS {
										break
									}
									time.Sleep(1 * time.Millisecond)
								}
								mutex.Lock()
								if i < len(tli.Chains) {
									tli.Chains[i].PlayNote(s.Notes, false)
								}
								mutex.Unlock()
							}(step)
						}
					}
					tli.Chains[i].TimePosition = timePosition
				}
				mutex.Unlock()
			}
		}
	}()
}
