package parser

import (
	"testing"
)

func TestParseMidi(t *testing.T) {
	// table driven tests
	tests := []struct {
		midiString string
		midiNear   int
		expected   []Note
	}{
		{"c", 71, []Note{{Midi: 72, Name: "c5"}}},
		{"c6", 20, []Note{{Midi: 84, Name: "c6"}}},
		{"c", 62, []Note{{Midi: 60, Name: "c4"}}},
		{"d", 32, []Note{{Midi: 26, Name: "d1"}}},
		{"f#3", 32, []Note{{Midi: 54, Name: "f#3"}}},
		{"g7", 100, []Note{{Midi: 103, Name: "g7"}}},
		{"gb", 100, []Note{{Midi: 103, Name: "g7"}, {Midi: 107, Name: "b7"}}},
		{"g♭c", 100, []Note{{Midi: 102, Name: "f#7"}, {Midi: 96, Name: "c7"}}},
		{"c4eg", 52, []Note{
			{Midi: 60, Name: "c4"},
			{Midi: 64, Name: "e4"},
			{Midi: 67, Name: "g4"},
		}},
		{"ceg6", 52, []Note{
			{Midi: 48, Name: "c3"},
			{Midi: 52, Name: "e3"},
			{Midi: 91, Name: "g6"},
		}},
	}
	for _, test := range tests {
		notes, err := ParseMidi(test.midiString, test.midiNear)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if len(notes) != len(test.expected) {
			t.Errorf("expected %v, got %v", test.expected, notes)
			continue
		}
		for i, note := range notes {
			if note.Midi != test.expected[i].Midi || note.Name != test.expected[i].Name {
				t.Errorf("test: %s (%d), expected %v, got %v", test.midiString, test.midiNear, test.expected, notes)
				break
			}
		}
	}
}
