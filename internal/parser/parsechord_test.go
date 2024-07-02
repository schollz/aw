package parser

import (
	"testing"

	log "github.com/schollz/logger"
)

func TestParseChord(t *testing.T) {
	log.SetLevel("trace")

	// table driven tests
	tests := []struct {
		chordString string
		midiNear    int
		expected    []Note
	}{
		{"Dm7/A;3", 60, []Note{
			{Midi: 57, Name: "a3"},
			{Midi: 60, Name: "c4"},
			{Midi: 62, Name: "d4"},
			{Midi: 65, Name: "f4"},
		}},
		{"Cm", 32, []Note{
			{Midi: 24, Name: "c1"},
			{Midi: 27, Name: "d#1"},
			{Midi: 31, Name: "g1"},
		}},
		{"Amaj7", 70, []Note{
			{Midi: 69, Name: "a4"},
			{Midi: 73, Name: "c#5"},
			{Midi: 76, Name: "e5"},
			{Midi: 80, Name: "g#5"},
		}},
		{"G", 70, []Note{
			{Midi: 67, Name: "g4"},
			{Midi: 71, Name: "b4"},
			{Midi: 74, Name: "d5"},
		}},
		{"G7", 70, []Note{
			{Midi: 67, Name: "g4"},
			{Midi: 71, Name: "b4"},
			{Midi: 74, Name: "d5"},
			{Midi: 77, Name: "f5"},
		}},
		{"Gmaj7", 70, []Note{
			{Midi: 67, Name: "g4"},
			{Midi: 71, Name: "b4"},
			{Midi: 74, Name: "d5"},
			{Midi: 78, Name: "f#5"},
		}},
		{"Gmaj7/F#", 70, []Note{
			{Midi: 66, Name: "f#4"},
			{Midi: 67, Name: "g4"},
			{Midi: 71, Name: "b4"},
			{Midi: 74, Name: "d5"},
		}},
		{"Cm", 70, []Note{
			{Midi: 60, Name: "c4"},
			{Midi: 63, Name: "d#4"},
			{Midi: 67, Name: "g4"},
		}},
	}

	for _, test := range tests {
		midiNotes, err := ParseChord(test.chordString, test.midiNear)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if len(midiNotes) != len(test.expected) {
			t.Errorf("'%s': Expected %d notes, got %d", test.chordString, len(test.expected), len(midiNotes))
		}
		for i, note := range midiNotes {
			if note.Midi != test.expected[i].Midi {
				t.Errorf("'%s': Expected %d, got %d", test.chordString, test.expected[i].Midi, note.Midi)
			}
			if note.Name != test.expected[i].Name {
				t.Errorf("'%s': Expected %s, got %s", test.chordString, test.expected[i].Name, note.Name)
			}
		}
	}

}
