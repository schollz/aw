package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseChain(expr string) ([]string, error) {
	var stack []string
	var currentToken strings.Builder
	if !strings.Contains(expr, "chain") {
		return nil, fmt.Errorf("chain not found")
	}
	expr = strings.Replace(expr, "chain", "", -1)
	expr = strings.TrimSpace(expr)

	pushCurrentToken := func() {
		if currentToken.Len() > 0 {
			stack = append(stack, currentToken.String())
			currentToken.Reset()
		}
	}

	for _, char := range expr {
		if char == '[' || char == ']' {
			pushCurrentToken()
			stack = append(stack, string(char))
			continue
		}

		if char == ' ' {
			pushCurrentToken()
			continue
		}

		if char == '*' {
			pushCurrentToken()
			stack = append(stack, "*")
			continue
		}

		currentToken.WriteRune(char)
	}
	pushCurrentToken()

	var evaluate func([]string) ([]string, error)
	evaluate = func(tokens []string) ([]string, error) {
		var result []string
		i := 0

		for i < len(tokens) {
			token := tokens[i]

			if token == "[" {
				closeIdx := findClosingParen(tokens[i+1:])
				if closeIdx == -1 {
					return nil, fmt.Errorf("unmatched parentheses")
				}
				innerResult, err := evaluate(tokens[i+1 : i+1+closeIdx])
				if err != nil {
					return nil, err
				}

				if i+closeIdx+2 < len(tokens) && tokens[i+closeIdx+2] == "*" {
					if multiplier, err := strconv.Atoi(tokens[i+closeIdx+3]); err == nil {
						expandedResult := []string{}
						for j := 0; j < multiplier; j++ {
							expandedResult = append(expandedResult, innerResult...)
						}
						result = append(result, expandedResult...)
						i += closeIdx + 4
						continue
					}
				}

				result = append(result, innerResult...)
				i += closeIdx + 2
				continue
			}

			if token != "*" {
				result = append(result, token)
				i++
				continue
			}

			if i > 0 && i+1 < len(tokens) {
				prevToken := result[len(result)-1]
				nextToken := tokens[i+1]

				if count, err := strconv.Atoi(nextToken); err == nil {
					result = result[:len(result)-1] // Remove last element as it will be multiplied
					for j := 0; j < count; j++ {
						result = append(result, prevToken)
					}
					i += 2 // Skip next token as it's part of multiplication
					continue
				}
			}

			result = append(result, token)
			i++
		}

		return result, nil
	}

	result, err := evaluate(stack)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func findClosingParen(tokens []string) int {
	depth := 1

	for i, token := range tokens {
		if token == "[" {
			depth++
		} else if token == "]" {
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}
