package parser

import (
	"fmt"
	"testing"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func TestArpeggio(t *testing.T) {
	log.SetLevel("debug")
	tests := []struct {
		line     string
		expected string
	}{
		{"Cm(rud4)", "c4 d#4 c4 g3 d#3"},
		{"F(ru4d2u4)", "f4 a4 c5 f5 a5 f5 c5 f5 a5 c6"},
		{"F(ru4d2u4,v4)", "f4(v4) a4(v4) c5(v4) f5(v4) a5(v4) f5(v4) c5(v4) f5(v4) a5(v4) c6(v4)"},
		{"a b c", "a b c"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("line(%s)", test.line), func(t *testing.T) {
			tokens, _ := TokenizeLineString(test.line)
			tokens, err := RetokenizeArpeggioArgument(tokens)
			if err != nil {
				log.Tracef("Error parsing expression: %v", err)
				if test.expected != "error" {
					t.Fatalf("Error parsing expression: %v", err)
				}
			} else {
				result := TokenExpandToLine(tokens)
				if result != test.expected {
					t.Fatalf("\n\t%s -->\n\t%v != %v", test.line, result, test.expected)
				}
			}
		})
	}
}

func TestTokens(t *testing.T) {
	log.SetLevel("debug")
	tests := []struct {
		line     string
		expected []string
	}{
		{"[a(b,c )]", []string{"[", "a(b,c )", "]"}},
		{"a [a(b,c )] c", []string{"a", "[", "a(b,c )", "]", "c"}},
		{"[a(b, c)]", []string{"[", "a(b, c)", "]"}},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("line(%s)", test.line), func(t *testing.T) {
			tokens, _ := TokenizeLineString(test.line)
			assert.Equal(t, test.expected, tokens)
		})
	}
}
func TestLine(t *testing.T) {
	log.SetLevel("debug")
	tests := []struct {
		line     string
		expected string
	}{
		{"Cm7 d e", "Cm7 d e"},
		{"a", "a"},
		{"a*8", "a a a a a a a a"},
		{"a(v) [b c] d", "a(v) _ b c d _"},
		{"a [b c]*2 d ~", "a _ _ _ b c b c d _ _ _ ~ _ _ _"},
		{"a [b c] d [e f]", "a _ b c d _ e f"},
		{"a(x4, v4) [b c] d [e f]", "a(x4, v4) _ b c d _ e f"},
		{"C(ru3,v4) f4", "c4(v4) e4(v4) g4(v4) f4"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("line(%s)", test.line), func(t *testing.T) {
			line := test.line
			line = ExpandMultiplication(line)
			tokens, err := TokenizeLineString(line)
			log.Debugf("tokens: %v", tokens)
			if err != nil {
				log.Debugf("Error parsing expression: %v", err)
				if test.expected != "error" {
					t.Fatalf("Error parsing expression: %v", err)
				}
			} else {
				tokens, _ = RetokenizeArpeggioArgument(tokens)
				log.Debugf("tokens: %v", tokens)
				result := TokenExpandToLine(tokens)
				if result != test.expected {
					t.Fatalf("\n\t%s -->\n\t%v != %v", test.line, result, test.expected)
				}
			}
		})
	}
}
