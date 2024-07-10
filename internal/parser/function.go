package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func (f Function) GetFloatPlace(name string, place int) (val float64, err error) {
	for _, arg := range f.Args {
		if arg.Name == name {
			val, err = strconv.ParseFloat(arg.Value, 64)
			return
		}
	}
	if place < len(f.Args) {
		val, err = strconv.ParseFloat(f.Args[place].Value, 64)
		return
	}
	err = fmt.Errorf("could not find argument %s or place %d", name, place)

	return
}

func (f Function) GetFloat(name string) (val float64, err error) {
	for _, arg := range f.Args {
		if arg.Name == name {
			val, err = strconv.ParseFloat(arg.Value, 64)
			return
		}
	}
	err = fmt.Errorf("could not find argument %s", name)
	return
}
func (f Function) GetStringPlace(name string, place int) (val string, err error) {
	for _, arg := range f.Args {
		if arg.Name == name {
			val = arg.Value
			return
		}
	}
	if place < len(f.Args) {
		val = f.Args[place].Value
		return
	}
	err = fmt.Errorf("could not find argument %s or place %d", name, place)
	return
}

func (f Function) GetInt(name string) (val int, err error) {
	for _, arg := range f.Args {
		if arg.Name == name {
			val, err = strconv.Atoi(arg.Value)
			return
		}
	}
	err = fmt.Errorf("could not find argument %s", name)
	return
}

func (f Function) GetIntPlace(name string, place int) (val int, err error) {
	for _, arg := range f.Args {
		if arg.Name == name {
			val, err = strconv.Atoi(arg.Value)
			return
		}
	}
	if place < len(f.Args) {
		val, err = strconv.Atoi(f.Args[place].Value)
		return
	}
	err = fmt.Errorf("could not find argument %s or place %d", name, place)
	return
}

func SplitArgFloat(arg string) []float64 {
	// arg in form (0.1,2.1,0.3)
	arg = strings.Trim(arg, "()")
	parts := strings.Split(arg, ",")
	var vals []float64
	for _, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseFloat(part, 64)
		if err == nil {
			vals = append(vals, val)
		}
	}
	return vals
}

func SplitArg(arg string) []int {
	// arg in form (1,2,3)
	arg = strings.Trim(arg, "()")
	parts := strings.Split(arg, ",")
	var vals []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.Atoi(part)
		if err == nil {
			vals = append(vals, val)
		}
	}
	return vals
}

type Function struct {
	Name string
	Args []Arg
}

type Arg struct {
	Name  string
	Value string
}

func ParseFunction(text string) (f Function, err error) {
	// remove all spaces
	text = strings.ReplaceAll(text, " ", "")
	openParenIndex := strings.Index(text, "(")
	closeParenIndex := strings.LastIndex(text, ")")

	if openParenIndex == -1 || closeParenIndex == -1 || openParenIndex > closeParenIndex {
		f.Name = text
		return f, nil
	}

	f.Name = text[:openParenIndex]
	argsStr := text[openParenIndex+1 : closeParenIndex]

	if argsStr == "" {
		return f, nil // No arguments
	}

	f.Args, err = parseArgs(argsStr)
	if err != nil {
		return f, err
	}

	return f, nil
}

func parseArgs(argsStr string) ([]Arg, error) {
	var args []Arg
	var currentArg string
	openParenCount := 0

	for i := 0; i < len(argsStr); i++ {
		char := argsStr[i]
		if char == '(' {
			openParenCount++
		} else if char == ')' {
			openParenCount--
		}

		if char == ',' && openParenCount == 0 {
			arg, err := parseArg(currentArg)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			currentArg = ""
		} else {
			currentArg += string(char)
		}
	}

	if currentArg != "" {
		arg, err := parseArg(currentArg)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

func parseArg(argStr string) (Arg, error) {
	arg := Arg{}
	if strings.Contains(argStr, "=") {
		parts := strings.SplitN(argStr, "=", 2)
		arg.Name = parts[0]
		arg.Value = parts[1]
	} else {
		arg.Value = argStr
	}
	return arg, nil
}
