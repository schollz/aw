package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func findMaxPrefix(a string, b string) string {
	i := 0
	for i < len(a) && i < len(b) {
		if a[i] != b[i] {
			break
		}
		i++
	}
	return a[:i]
}

func exactMatch(n string) (note Note, ok bool) {
	for _, m := range noteDB {
		for _, noteFullName := range append(m.NamesOther, strings.ToLower(m.NameSharp)) {
			if n == noteFullName {
				return Note{Midi: m.MidiValue, NameOriginal: n, Name: strings.ToLower(m.NameSharp)}, true
			}
		}
	}
	return
}
func ParseMidi(midiString string, midiNear int) (notes []Note, err error) {
	// can be a single midi note like "c" in which case we need to find the closest note to midiNear
	// or can be a single note like "c4" in which case we want an exact match
	// or can be a sequence of notes like "c4eg" in which case we want to need to split them
	midiString = strings.ToLower(midiString)

	// check if split if it has multiple of any letter [a-g]
	noteStrings := []string{}
	lastAdded := 0
	for i := 1; i < len(midiString); i++ {
		if midiString[i] >= 'a' && midiString[i] <= 'g' {
			noteStrings = append(noteStrings, midiString[lastAdded:i])
			lastAdded = i
		}
	}
	if lastAdded != len(midiString) {
		noteStrings = append(noteStrings, midiString[lastAdded:])
	}
	// log.Debugf("%s' -> %v", midiString, noteStrings)

	// convert '#' to 's'
	for i, n := range noteStrings {
		noteStrings[i] = strings.Replace(n, "#", "s", -1)
	}

	// convert '♭' to 'b'
	for i, n := range noteStrings {
		noteStrings[i] = strings.Replace(n, "♭", "b", -1)
	}

	notes = make([]Note, len(noteStrings))
	for i, n := range noteStrings {
		if note, ok := exactMatch(n); ok {
			notes[i] = note
			midiNear = note.Midi
		} else {
			// find closes to midiNear
			newNote := Note{Midi: 300, Name: ""}
			closestDistance := math.Inf(1)
			for _, m := range noteDB {
				for octave := -1; octave <= 8; octave++ {
					for _, noteFullName := range append(m.NamesOther, strings.ToLower(m.NameSharp)) {
						noteName := findMaxPrefix(n, noteFullName)
						if noteName != "" && (noteName == noteFullName || (noteName+strconv.Itoa(octave)) == noteFullName) {
							if math.Abs(float64(m.MidiValue-midiNear)) < closestDistance {
								closestDistance = math.Abs(float64(m.MidiValue - midiNear))
								newNote = Note{Midi: m.MidiValue, NameOriginal: noteName, Name: strings.ToLower(m.NameSharp)}
							}
						}
					}
				}
			}
			if newNote.Midi != 300 {
				notes[i] = newNote
				midiNear = newNote.Midi
			} else {
				err = fmt.Errorf("parsemidi could not parse %s", n)
				return
			}

		}
	}

	return

}
