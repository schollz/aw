package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type Function struct {
	Name string
	Args []Arg
}

type Arg struct {
	Name  string
	Value string
}

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

func (f Function) GetInt(name string, place int) (val int, err error) {
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

	argsParts := strings.Split(argsStr, ",")
	for _, part := range argsParts {
		arg := Arg{}
		if strings.Contains(part, "=") {
			parts := strings.SplitN(part, "=", 2)
			arg.Name = parts[0]
			arg.Value = parts[1]
		} else {
			arg.Value = part
		}
		f.Args = append(f.Args, arg)
	}

	return f, nil
}
