package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	log "github.com/schollz/logger"
)

// RetokenizeArpeggioDecorator takes a line of notes and/or chords with decorators and expands it
// for example Cm7 rud4 will expand to an arpeggio that goes up 1 and down 4 for a total of 5 steps
func RetokenizeArpeggioArgument(tokens []string) (newTokens []string, err error) {
	oldTokens := make([]string, len(tokens))
	copy(oldTokens, tokens)
	// gather tokens
	for i, token := range tokens {
		_ = i
		var fn Function
		fn, err = ParseFunction(token)
		if err != nil {
			log.Error(err)
			return
		}
		arpPiece := ""
		hasArp := false
		newArgs := []string{}
		for _, arg := range fn.Args {
			if strings.HasPrefix(arg.Value, "r") {
				hasArp = true
				arpPiece = arg.Value
			} else {
				newArgs = append(newArgs, arg.Value)
			}
		}
		if !hasArp {
			continue
		}
		newFnArgs := "(" + strings.Join(newArgs, ",") + ")"
		if len(newArgs) == 0 {
			newFnArgs = ""
		}
		log.Debug(newFnArgs)
		// get the notes
		notes, errParse := ParseChord(fn.Name, 60)
		if errParse != nil {
			notes, errParse = ParseMidi(fn.Name, 60)
		}
		if errParse != nil {
			err = fmt.Errorf("cannot parse %s: %v", token, errParse)
			return
		}
		log.Debugf("notes: %v", notes)
		notei := 0
		var noteString strings.Builder
		for j, piece := range tokenizeLetterNumberes(arpPiece) {
			if j == 0 {
				continue
			}
			// get the number
			num := 1
			if len(piece) > 1 {
				num, err = strconv.Atoi(piece[1:])
				if err != nil {
					num = 1
				}
			}
			for k := 0; k < num; k++ {
				// add a note
				noteIndex := notei
				for noteIndex < 0 {
					noteIndex += len(notes)
				}
				newNote := notes[noteIndex%len(notes)]
				octave := notei / len(notes)
				if notei < 0 {
					octave--
				}
				newNote = NoteAdd(newNote, octave*12)
				noteString.WriteString(strings.ToLower(newNote.Name) + newFnArgs + " ")
				switch piece[0] {
				case 'u':
					notei++
				case 'd':
					notei--
				}
			}
		}
		tokens[i] = noteString.String()
	}
	newTokens = tokens
	log.Debugf("RetokenizeArpeggioDecorator: %v->%v", oldTokens, newTokens)
	return
}

func TokenizeLineString(s string) (tokens []string, err error) {
	// gather tokens
	tokens = tokenizeFromGrouping(s)

	// check to see whether parentheses match
	err = doParenthesesMatch(tokens)
	return
}

func tokenizeFromGrouping(s string) (tokens []string) {
	var i int
	for i < len(s) {
		switch s[i] {
		case '[':
			tokens = append(tokens, string(LEFT_GROUP))
			i++
		case ']':
			tokens = append(tokens, string(RIGHT_GROUP))
			i++
		case ' ':
			i++
		default:
			if s[i] == '(' {
				// Error: we should not encounter '(' without a preceding function name
				i++
			} else {
				// Find the next space or parenthesis to capture the token
				start := i
				for i < len(s) && s[i] != ' ' && s[i] != '(' && s[i] != ')' && s[i] != '[' && s[i] != ']' {
					i++
				}
				if i < len(s) && s[i] == '(' {
					// We have encountered a function call
					startFunc := start
					i++ // move past the '('
					stack := 1
					for i < len(s) && stack > 0 {
						if s[i] == '(' {
							stack++
						} else if s[i] == ')' {
							stack--
						}
						i++
					}
					tokens = append(tokens, s[startFunc:i])
				} else {
					tokens = append(tokens, s[start:i])
				}
			}
		}
	}
	return
}

// TokenExpandToLine takes a line of notes and/or chords with decorators and expands it
// for example Cm7 rud (d e f g) will expand to
// Cm7-rud - - - d e f g
// where the decorators are attached to the enttity and the entities are
// given sustains where nessecary (Cm7 is sustained)
func TokenExpandToLine(tokens []string) (expanded string) {

	currentTokenValues := make([]float64, len(tokens))
	for i := range currentTokenValues {
		currentTokenValues[i] = 1
	}
	tokenValues := determineTokenValue(tokens, currentTokenValues, 0, len(tokens))
	// print values of non-parentheses tokens
	minValue := math.Inf(1)
	for i, token := range tokens {
		if token != LEFT_GROUP && token != RIGHT_GROUP {
			if tokenValues[i] < minValue {
				minValue = tokenValues[i]
			}
		}
	}
	// now expand it by building a string
	sb := strings.Builder{}
	for i, token := range tokens {
		if token != LEFT_GROUP && token != RIGHT_GROUP {
			repetitions := int(math.Round(tokenValues[i] / minValue))
			sb.WriteString(token + " ")
			for j := 1; j < repetitions; j++ {
				sb.WriteString(HOLD + " ")
			}
		}
	}
	expanded = strings.TrimSpace(sb.String())
	expanded = strings.Join(strings.Fields(expanded), " ")
	return
}

func doParenthesesMatch(tokens []string) (err error) {
	// check to see whether parentheses match
	// and if not, return a string showing where the error is
	depth := 0
	lastDepthIncrease := 0
	for i, token := range tokens {
		if token == LEFT_GROUP {
			depth++
			lastDepthIncrease = i
		} else if token == RIGHT_GROUP {
			depth--
			if depth < 0 {
				sb := strings.Builder{}
				sb.WriteString("\n")
				sb.WriteString(strings.Join(tokens, " ") + "\n")
				for j := 0; j < i; j++ {
					sb.WriteString("  ")
				}
				sb.WriteString("^ parentheses do not match")
				err = fmt.Errorf(sb.String())
				return
			}
		}
	}
	if depth != 0 {
		sb := strings.Builder{}
		sb.WriteString("\n")
		sb.WriteString(strings.Join(tokens, " ") + "\n")
		for i := 0; i < lastDepthIncrease; i++ {
			sb.WriteString(" ")
		}
		sb.WriteString("^ parentheses do not match")
		err = fmt.Errorf(sb.String())
	}

	return
}

func determineTokenValue(tokens []string, currentTokenValues []float64, start int, stop int) (tokenValues []float64) {
	depth := 0
	tokenLocations := [][2]int{} // start, end
	tokenValues = make([]float64, len(tokens))
	copy(tokenValues, currentTokenValues)
	for i := start; i < stop; i++ {
		token := tokens[i]
		if token == LEFT_GROUP {
			if depth == 0 {
				tokenLocations = append(tokenLocations, [2]int{i, -1})
			}
			depth++
		} else if token == RIGHT_GROUP {
			depth--
			if depth == 0 {
				tokenLocations[len(tokenLocations)-1][1] = i
			}
		}
	}
	for _, loc := range tokenLocations {
		tokenValues = determineTokenValue(tokens, tokenValues, loc[0]+1, loc[1])
	}

	numEntities := countEntities(tokens[start:stop])
	valuePerEntity := 1.0 / float64(numEntities)
	for i := start; i < stop; i++ {
		tokenValues[i] = valuePerEntity * tokenValues[i]
	}
	return
}

func countEntities(tokens []string) (entities int) {
	depth := 0
	for _, token := range tokens {
		if token == LEFT_GROUP {
			if depth == 0 {
				entities++
			}
			depth++
		} else if token == RIGHT_GROUP {
			depth--
		} else if depth == 0 {
			entities++
		}
	}
	return
}

func tokenizeLetterNumberes(s string) (tokens []string) {
	// take a string with letters and numbers and split it into tokens
	// where the letter determines the start of a new token
	for i := 0; i < len(s); i++ {
		if i == 0 {
			tokens = append(tokens, string(s[i]))
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			tokens[len(tokens)-1] += string(s[i])
		} else {
			tokens = append(tokens, string(s[i]))
		}
	}
	return
}
