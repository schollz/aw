package parser

import (
	"testing"

	log "github.com/schollz/logger"
)

func TestParseFunction(t *testing.T) {
	log.SetLevel("debug")
	tests := []struct {
		text string
		want Function
	}{
		{"crow( 1)", Function{Name: "crow", Args: []Arg{{Value: "1"}}}},
		{"F( ru4d2u4, v4)", Function{Name: "F", Args: []Arg{{Value: "ru4d2u4"}, {Value: "v4"}}}},
		{"run(v=4, d=2)", Function{Name: "run", Args: []Arg{{Name: "v", Value: "4"}, {Name: "d", Value: "2"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := ParseFunction(tt.text)
			if err != nil {
				t.Errorf("ParseFunction() error = %v", err)
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("ParseFunction() got = %v, want %v", got.Name, tt.want.Name)
			}
			if len(got.Args) != len(tt.want.Args) {
				t.Errorf("ParseFunction() got = %v, want %v", got.Args, tt.want.Args)
			}
			for i := range got.Args {
				if got.Args[i].Name != tt.want.Args[i].Name {
					t.Errorf("ParseFunction() got = %v, want %v", got.Args[i].Name, tt.want.Args[i].Name)
				}
				if got.Args[i].Value != tt.want.Args[i].Value {
					t.Errorf("ParseFunction() got = %v, want %v", got.Args[i].Value, tt.want.Args[i].Value)
				}
			}
		})
	}

}
