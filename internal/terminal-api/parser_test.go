package terminalApi

import (
	"reflect"
	"strings"
	"testing"
)

func testParser(t *testing.T) Parser {
	t.Helper()
	parser, err := NewParser(Command{
		Name: "trustpkg-cli",
		Flags: []Flag{
			{Name: "output", Short: "o", Type: ListFlag, Persistent: true},
			{Name: "should-use", Type: BoolFlag, Persistent: true},
		},
		Commands: []Command{{Name: "generate", Description: "Generate files", Flags: []Flag{
			{Name: "name", Short: "n", Type: StringFlag, StringDefault: "default"},
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return parser
}

func TestParseInvocation(t *testing.T) {
	parser := testParser(t)
	result, err := parser.Parse([]string{"-o", "ffdfdf,test", "--should-use", "true", "generate", "xxfdfd", "-o", "third", "-n=hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Command, []string{"trustpkg-cli", "generate"}) ||
		!reflect.DeepEqual(result.Arguments, []string{"xxfdfd"}) {
		t.Fatalf("unexpected invocation: %+v", result)
	}
	if output := result.Flags["output"]; !reflect.DeepEqual(output.List, []string{"ffdfdf", "test", "third"}) || !output.Changed {
		t.Fatalf("unexpected output: %+v", output)
	}
	if value := result.Flags["should-use"]; !value.Bool || !value.Changed {
		t.Fatalf("unexpected bool: %+v", value)
	}
	if value := result.Flags["name"]; value.String != "hello" || !value.Changed {
		t.Fatalf("unexpected string: %+v", value)
	}
	result, err = parser.Parse([]string{"generate", "--", "--name", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Arguments, []string{"--name", "x"}) || result.Flags["name"].String != "default" || result.Flags["name"].Changed {
		t.Fatalf("unexpected arguments or default: %+v", result)
	}
}

func TestParseErrorsAndHelp(t *testing.T) {
	parser := testParser(t)
	for _, args := range [][]string{
		{"unknown"}, {"generate", "--unknown"}, {"--should-use"},
		{"--should-use", "invalid"}, {"--output", `"unterminated`},
	} {
		if _, err := parser.Parse(args); err == nil {
			t.Errorf("expected error for %q", args)
		}
	}
	result, err := parser.Parse([]string{"generate", "--help"})
	if err != nil || !result.HelpRequested || !strings.Contains(result.Help, "--name") {
		t.Fatalf("unexpected help: %+v, %v", result, err)
	}
	result, err = parser.Parse([]string{"help", "generate"})
	if err != nil || !result.HelpRequested || !strings.Contains(result.Help, "--name") {
		t.Fatalf("unexpected help command: %+v, %v", result, err)
	}
	result, err = parser.Parse([]string{"--should-use=false"})
	if err != nil || result.Flags["should-use"].Bool || !result.Flags["should-use"].Changed {
		t.Fatalf("unexpected false flag: %+v, %v", result, err)
	}
}

func TestRequiredFlag(t *testing.T) {
	parser, err := NewParser(Command{Name: "cli", Flags: []Flag{{Name: "name", Type: StringFlag, Required: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse(nil); err == nil {
		t.Fatal("expected missing required flag error")
	}
	if _, err := parser.Parse([]string{"--name", "value"}); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidSchema(t *testing.T) {
	for _, spec := range []Command{
		{Name: "cli", Flags: []Flag{{Name: "bad", Type: "invalid"}}},
		{Name: "cli", Flags: []Flag{{Name: "a", Type: StringFlag}, {Name: "a", Type: BoolFlag}}},
		{Name: "cli", Flags: []Flag{{Name: "a", Short: "x", Type: StringFlag, Persistent: true}},
			Commands: []Command{{Name: "sub", Flags: []Flag{{Name: "b", Short: "x", Type: BoolFlag}}}}},
	} {
		if _, err := NewParser(spec); err == nil {
			t.Errorf("expected invalid schema error for %+v", spec)
		}
	}
}
