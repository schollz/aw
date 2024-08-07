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
	"github.com/schollz/gomidi"
	log "github.com/schollz/logger"
)

var crows crow.Murder
var mutex sync.Mutex
var midiDevices map[string]gomidi.Device

func init() {
	midiDevices = make(map[string]gomidi.Device)
}

type TLI struct {
	TimePosition   []int64 `json:"time_position,omitempty"`
	Chains         []Chain `json:"chains"`
	ChainsRendered []Chain `json:"rendered"`
	Loops          []Loop  `json:"loops"`
	Params         Params  `json:"params"`
	Playing        bool    `json:"playing"`
}

type Params struct {
	IsSet    int `json:"isset"`
	Tempo    int `json:"tempo"`
	Gate     int `json:"gate"`
	Velocity int `json:"velocity"`
	Index    int `json:"index"`
}

const (
	TempoSet = 1 << iota
	GateSet
	VelocitySet
	IndexSet
)

func (p *Params) Set(param int, val int) {
	p.IsSet |= param
	switch param {
	case TempoSet:
		p.Tempo = val
	case GateSet:
		p.Gate = val
	case VelocitySet:
		p.Velocity = val
	case IndexSet:
		p.Index = val
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
		sb.WriteString(fmt.Sprintf("\ntie '%s':", p.Name))
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

func PlayNote(notes []Note, on bool, outFns []Function) (err error) {
	for _, out := range outFns {
		log.Debugf("[%+v] note %v: %+v", out, on, notes)
		switch out.Name {
		case "midi":
			var output string
			output, err = out.GetStringPlace("name", 0)
			if err != nil {
				log.Error(err)
				return
			}
			channel, _ := out.GetIntPlace("ch", 1)
			log.Tracef("midi out: %s %d", output, channel)
			for _, note := range notes {
				if on {
					midiDevices[output].NoteOn(uint8(channel), uint8(note.Midi), 120)
				} else {
					midiDevices[output].NoteOff(uint8(channel), uint8(note.Midi))
				}
			}
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
						crows.SetVoltage(output+j, float64(note.Midi-12.0)/12.0)
					}
					if crows.UseEnv[output] > 0 {
						crows.On(crows.UseEnv[output], on)
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

func New(text string) (tli *TLI, err error) {
	tli = new(TLI)
	tli.Params = Params{Tempo: 120}
	tli.TimePosition = make([]int64, 128)
	err = tli.ParseText(text)
	if err != nil {
		log.Error(err)
		return
	}
	err = tli.Render()
	if err != nil {
		log.Error(err)
		return
	}
	b, _ := json.Marshal(tli.Chains)
	json.Unmarshal(b, &tli.ChainsRendered)

	return
}

func (tli *TLI) Update(text string) (err error) {
	tliTest, err := New(text)
	if err != nil {
		log.Error(err)
		return
	}
	// copy over the rendered chains
	mutex.Lock()
	tliTest.Playing = tli.Playing
	for i, v := range tli.TimePosition {
		tliTest.TimePosition[i] = v
	}
	b, _ := json.Marshal(tliTest)
	json.Unmarshal(b, &tli)
	mutex.Unlock()
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
		if strings.HasPrefix(line, "#") {
			// skip comments
			continue
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "run") {
			fnFinish()
			state = StateLoop
			loop.Name = strings.Split(line, " ")[1]
			continue
		} else if strings.HasPrefix(line, "tie") {
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
					chain.Outs = append(chain.Outs, strings.TrimSpace(strings.TrimPrefix(line, "out")))
				}
			case StateSet:
				// parse set
				if strings.HasPrefix(line, "bpm") {
					val, errParse := strconv.Atoi(strings.Fields(line)[1])
					if errParse == nil {
						tli.Params.Set(TempoSet, val)
					}
				}
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
	line = SanitizeLine(line)
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
				velocity, errParse := strconv.Atoi(decorator[1:])
				if errParse == nil {
					step.Params.Set(VelocitySet, velocity)
				}
			} else if strings.HasPrefix(decorator, "h") {
				gate, errParse := strconv.Atoi(decorator[1:])
				if errParse == nil {
					step.Params.Set(GateSet, gate)
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
		lastGate := 95
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
			case "midi":
				name, errFind := fn.GetStringPlace("output", 0)
				if errFind == nil {
					if _, ok := midiDevices[name]; !ok {
						var err error
						midiDevices[name], err = gomidi.New(name)
						if err != nil {
							log.Error(err)
						} else {
							midiDevices[name].Open()
						}
					}
				}
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
					var output int
					output, _ = fn.GetIntPlace("output", 0)
					if output > 0 {
						crows.UseEnv[output], _ = fn.GetInt("env")

						if val, err = fn.GetFloat("slew"); err == nil {
							crows.SetSlew(output, val)
						}

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
	log.Tracef("OutFns: %+v", c.OutFns)

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
	if len(tli.ChainsRendered) > 0 {
		go tli.run()
	}
}

func setCrowAdsr(chain Chain, step Step, arg Arg) {
	log.Trace("[setCrowAdsr] start")
	for _, out := range chain.OutFns {
		if out.Name == "crow" {
			log.Trace("[setCrowAdsr] have crow out")
			// set adsr
			output, errInt := out.GetIntPlace("output", 0)
			log.Tracef("[setCrowAdsr] output: %d", output)
			log.Trace(errInt)
			log.Trace(crows.IsReady)

			if errInt == nil && crows.IsReady {
				vals := SplitArgFloat(arg.Value)
				log.Tracef("arg: %+v, vals: %+v", arg, vals)
				if len(vals) == 4 {
					log.Tracef("setting adsr: %+v", arg.Value)
					log.Tracef("step: %+v", step)
					for i, v := range vals {
						if i != 2 {
							vals[i] = v * float64(step.TimeDurationMicroseconds) / 1000000.0
						}
					}
					crows.SetADSR(output+1, crow.ADSR{Attack: vals[0], Decay: vals[1], Sustain: vals[2], Release: vals[3]})
				}
			}
		}
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
	for i := range tli.TimePosition {
		tli.TimePosition[i] = -1
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
				for i, chain := range tli.ChainsRendered {
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
						if (timePosition > step.TimeStartMicroseconds && tli.TimePosition[i] <= step.TimeStartMicroseconds) ||
							(timePosition < tli.TimePosition[i] && stepi == 0) {
							log.Info(timePosition, step.TimeStartMicroseconds, tli.TimePosition[i])
							if len(step.Arguments) > 0 {
								log.Infof("arguments: %+v", step.Arguments)
							}
							for _, arg := range step.Arguments {
								if strings.HasPrefix(arg.Value, "adsr") {
									arg.Value = strings.TrimPrefix(arg.Value, "adsr=")
									arg.Value = strings.TrimPrefix(arg.Value, "adsr")
									setCrowAdsr(chain, step, arg)
								}
							}
							PlayNote(step.Notes, true, chain.OutFns)
							go func(s Step, c Chain) {
								sleepMS := int64(math.Round(float64(s.TimeDurationMicroseconds) * float64(s.Params.Gate) / 100.0))
								sleepStart := hrtime.Now()
								for {
									if hrtime.Since(sleepStart).Microseconds() > sleepMS {
										break
									}
									time.Sleep(1 * time.Millisecond)
								}
								mutex.Lock()
								if i < len(tli.Chains) {
									PlayNote(s.Notes, false, c.OutFns)
								}
								if crows.NeedsFlush {
									crows.Flush()
								}
								mutex.Unlock()
							}(step, chain)
						}
					}
					tli.TimePosition[i] = timePosition
				}
				if crows.NeedsFlush {
					crows.Flush()
				}
				mutex.Unlock()
			}
		}
	}()
}
