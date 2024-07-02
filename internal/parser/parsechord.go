package parser

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/schollz/logger"
)

func ParseChord(chordString string, midiNear int) (result []Note, err error) {
	chordStringOriginal := chordString
	chordMatch := ""
	_ = chordMatch

	log.Tracef("chordString: %s", chordStringOriginal)

	octave := 4
	if midiNear != 0 {
		octave = midiNear/12 - 1
	}
	if strings.Contains(chordString, ";") {
		chordSplit := strings.Split(chordString, ";")
		chordString = chordSplit[0]
		if len(chordSplit) > 1 {
			octave, err = strconv.Atoi(chordSplit[1])
			if err != nil {
				return
			}
		}
	}
	log.Tracef("octave: %d", octave)
	log.Tracef("chordString: %s", chordString)

	transposeNote := ""
	if strings.Contains(chordString, "/") {
		chordSplit := strings.Split(chordString, "/")
		chordString = chordSplit[0]
		if len(chordSplit) > 1 {
			transposeNote = chordSplit[1]
		}
	}
	log.Tracef("transposeNote: %s", transposeNote)

	// find the root note name
	noteMatch := ""
	transposeNoteMatch := ""
	chordRest := ""
	for _, n := range notesAll {
		if transposeNote != "" && len(n) > len(transposeNoteMatch) {
			if strings.ToLower(n) == transposeNote || n == transposeNote {
				transposeNoteMatch = n
			}
		}
		if len(n) > len(noteMatch) {
			// check if has prefix
			if strings.HasPrefix(chordString, n) {
				noteMatch = n
				chordRest = chordString[len(n):]
			}
			if strings.HasPrefix(strings.ToLower(chordString), n) {
				noteMatch = n
				chordRest = chordString[len(n):]
			}
		}
	}
	if noteMatch == "" {
		err = fmt.Errorf("no chord found")
	}
	log.Tracef("noteMatch: %s", noteMatch)
	log.Tracef("chordRest: %s", chordRest)

	// convert to canonical sharp scale
	// e.g. Fb -> E, Gs -> G#
	for i, n := range notesScaleAcc1 {
		if noteMatch == n {
			noteMatch = notesScaleSharp[i]
			break
		}
	}
	for i, n := range notesScaleAcc2 {
		if noteMatch == n {
			noteMatch = notesScaleSharp[i]
			break
		}
	}
	for i, n := range notesScaleAcc3 {
		if noteMatch == n {
			noteMatch = notesScaleSharp[i]
			break
		}
	}
	if transposeNoteMatch != "" {
		for i, n := range notesScaleAcc1 {
			if transposeNoteMatch == n {
				transposeNoteMatch = notesScaleSharp[i]
				break
			}
		}
		for i, n := range notesScaleAcc2 {
			if transposeNoteMatch == n {
				transposeNoteMatch = notesScaleSharp[i]
				break
			}
		}
		for i, n := range notesScaleAcc3 {
			if transposeNoteMatch == n {
				transposeNoteMatch = notesScaleSharp[i]
				break
			}
		}
	}
	log.Tracef("noteMatch: %s", noteMatch)
	log.Tracef("transposeNoteMatch: %s", transposeNoteMatch)

	// find longest matching chord pattern
	chordMatch = "" // (no chord match is major chord)
	chordIntervals := "1P 3M 5P"
	for _, chordType := range dbChords {
		for i, chordPattern := range chordType {
			if i > 1 {
				if len(chordPattern) > len(chordMatch) && strings.ToLower(chordRest) == strings.ToLower(chordPattern) {
					chordMatch = chordPattern
					chordIntervals = chordType[0]
				}
			}
		}
	}
	log.Tracef("chordMatch for %s: %s", chordRest, chordMatch)
	log.Tracef("chordIntervals: %s", chordIntervals)

	// find location of root
	rootPosition := 0
	for i, n := range notesScaleSharp {
		if n == noteMatch {
			rootPosition = i
			break
		}
	}
	log.Tracef("rootPosition: %d", rootPosition)

	/** lua code
		-- find notes from intervals
	  whole_note_semitones={0,2,4,5,7,9,11,12}
	  notes_in_chord={}
	  for interval in string.gmatch(chord_intervals,"%S+") do
	    -- get major note position
	    major_note_position=(string.match(interval,"%d+")-1)%7+1
	    -- find semitones from root
	    semitones=whole_note_semitones[major_note_position]
	    -- adjust semitones based on interval
	    if string.match(interval,"m") then
	      semitones=semitones-1
	    elseif string.match(interval,"A") then
	      semitones=semitones+1
	    end
	    if self.debug then
	      print("interval: "..interval)
	      print("major_note_position: "..major_note_position)
	      print("semitones: "..semitones)
	      print("root_position+semitones: "..root_position+semitones)
	    end
	    -- get note in scale from root position
	    note_in_chord=self.notes_scale_sharp[root_position+semitones]
	    table.insert(notes_in_chord,note_in_chord)
	  end
	  **/

	// go code
	// find notes from intervals
	wholeNoteSemitones := []int{0, 2, 4, 5, 7, 9, 11, 12}
	notesInChord := []string{}
	for _, interval := range strings.Fields(chordIntervals) {
		// get major note position
		majorNotePosition, _ := strconv.Atoi(strings.TrimRight(interval, "mMAP"))
		majorNotePosition = ((majorNotePosition - 1) % 7) + 1
		// find semitones from root
		semitones := wholeNoteSemitones[majorNotePosition]
		// adjust semitones based on interval
		if strings.Contains(interval, "A") || strings.Contains(interval, "M") {
			semitones = semitones + 1
		}
		log.Trace("-------------------------")
		log.Tracef("interval: %s", interval)
		log.Tracef("majorNotePosition: %d", majorNotePosition)
		log.Tracef("semitones: %d", semitones)
		log.Tracef("rootPosition+semitones: %d", rootPosition+semitones)
		// get note in scale from root position
		noteInChord := notesScaleSharp[(rootPosition+semitones)-2+12]
		notesInChord = append(notesInChord, noteInChord)
	}
	log.Tracef("notesInChord: %v", notesInChord)
	log.Tracef("notesScaleSharp: %v", notesScaleSharp)

	// if tranposition, rotate until new root
	if transposeNoteMatch != "" {
		foundNote := false
		for i := 0; i < len(notesInChord); i++ {
			if notesInChord[0] == transposeNoteMatch {
				foundNote = true
				break
			}
			notesInChord = append(notesInChord[1:], notesInChord[0])
		}
		if !foundNote {
			notesInChord = append([]string{transposeNoteMatch}, notesInChord...)
		}
	}
	log.Tracef("notesInChord: %v", notesInChord)

	// go code
	// convert to midi
	midiNotesInChord := []int{}
	lastNote := 0
	for i, n := range notesInChord {
		for _, d := range noteDB {
			if d.MidiValue > lastNote &&
				(d.NameSharp == n+strconv.Itoa(octave) ||
					d.NameSharp == n+strconv.Itoa(octave+1) ||
					d.NameSharp == n+strconv.Itoa(octave+2) ||
					d.NameSharp == n+strconv.Itoa(octave+3)) {
				lastNote = d.MidiValue
				midiNotesInChord = append(midiNotesInChord, d.MidiValue)
				notesInChord[i] = d.NameSharp
				break
			}
		}
	}
	log.Tracef("midiNotesInChord: %v", midiNotesInChord)
	log.Tracef("notesInChord: %v", notesInChord)

	result = make([]Note, len(midiNotesInChord))
	for i, m := range midiNotesInChord {
		result[i] = Note{Midi: m, Name: strings.ToLower(notesInChord[i])}
	}
	log.Tracef("result: %v", result)

	return
}
