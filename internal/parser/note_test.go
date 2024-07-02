package parser

import (
	"fmt"
	"testing"
)

func TestNote(t *testing.T) {
	n := Note{Midi: 60, Name: "C"}
	n2 := NoteAdd(n, 1)
	if n2.Midi != 61 {
		t.Fatalf("expected %d, got %d", 61, n2.Midi)
	}
	fmt.Println(n2)

}
