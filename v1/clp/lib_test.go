package clp

import (
	"os"
	"reflect"
	"testing"
)

type exitPanic int

func expectExit(t *testing.T, fn func()) {
	t.Helper()

	oldExitProcess := exitProcess
	exitProcess = func(code int) {
		panic(exitPanic(code))
	}
	t.Cleanup(func() {
		exitProcess = oldExitProcess
	})

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected command parser to exit")
		}

		code, ok := recovered.(exitPanic)
		if !ok {
			panic(recovered)
		}

		if code != 1 {
			t.Fatalf("expected exit code 1, got %d", code)
		}
	}()

	fn()
}

func TestBoolEnvSurvivesAbsentFlag(t *testing.T) {
	t.Setenv("CLP_TEST_BOOL", "true")

	c := NewCmdParserFromArgs([]string{})

	if got := c.GetBool(false, "CLP_TEST_BOOL", c.Flags("--enabled"), ""); got != true {
		t.Fatalf("expected env var to set bool when CLI flag is absent, got %v", got)
	}
}

func TestBoolEnvFalseSurvivesAbsentFlag(t *testing.T) {
	t.Setenv("CLP_TEST_BOOL", "false")

	c := NewCmdParserFromArgs([]string{})

	if got := c.GetBool(true, "CLP_TEST_BOOL", c.Flags("--enabled"), ""); got != false {
		t.Fatalf("expected env var to clear bool when CLI flag is absent, got %v", got)
	}
}

func TestBareBooleanFlagEnablesDefaultFalse(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug"})

	if got := c.GetBool(false, "", c.Flags("--debug"), ""); got != true {
		t.Fatalf("expected bare bool flag to enable default false, got %v", got)
	}
}

func TestBooleanFlagAcceptsSeparatedFalseValue(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug", "0"})

	if got := c.GetBool(true, "", c.Flags("--debug"), ""); got != false {
		t.Fatalf("expected separated false value to clear bool, got %v", got)
	}
}

func TestValuesNonFlagArgsNegativeNumbersAndDoubleDash(t *testing.T) {
	c := NewCmdParserFromArgs([]string{
		"serve",
		"--port", "-1",
		"--name=api",
		"--",
		"--literal",
		"tail",
	})

	if got := c.GetInt(3000, "", c.Flags("--port"), ""); got != -1 {
		t.Fatalf("expected negative int value from CLI, got %v", got)
	}

	if got := c.GetString("web", "", c.Flags("--name"), ""); got != "api" {
		t.Fatalf("expected string value from equals flag, got %q", got)
	}

	wantArgs := []string{"serve", "--literal", "tail"}
	if !reflect.DeepEqual(c.NonFlagArgs, wantArgs) {
		t.Fatalf("expected non-flag args %v, got %v", wantArgs, c.NonFlagArgs)
	}
}

func TestMissingValueDoesNotConsumeNextFlag(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--name", "--debug"})

	meta, ok := c.FlagsMetaMap["--name"]
	if !ok {
		t.Fatal("expected --name metadata")
	}

	if !meta.NotEnoughArgs {
		t.Fatalf("expected --name to be marked as missing a value: %#v", meta)
	}

	if got := c.GetBool(false, "", c.Flags("--debug"), ""); got != true {
		t.Fatalf("expected following bare bool flag to remain parseable, got %v", got)
	}
}

func TestRepeatedFlagValuesAreTracked(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--name", "api", "--name=api"})

	meta, ok := c.FlagsMetaMap["--name"]
	if !ok {
		t.Fatal("expected --name metadata")
	}

	if got, want := meta.Values, []string{"api", "api"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected repeated values %v, got %v", want, got)
	}

	if got := c.GetString("web", "", c.Flags("--name"), ""); got != "api" {
		t.Fatalf("expected repeated matching value to parse, got %q", got)
	}
}

func TestStringCanBeSetToEmptyWithEquals(t *testing.T) {
	t.Setenv("CLP_TEST_NAME", "env")

	c := NewCmdParserFromArgs([]string{"--name="})

	if got := c.GetString("default", "CLP_TEST_NAME", c.Flags("--name"), ""); got != "" {
		t.Fatalf("expected empty string value from --name=, got %q", got)
	}
}

func TestNewCmdParserUsesOSArgsWithoutProgramName(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	os.Args = []string{"/tmp/program", "run", "--debug"}
	c := NewCmdParser()

	if got, want := c.NonFlagArgs, []string{"run"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected non-flag args %v, got %v", want, got)
	}

	if got := c.GetBool(false, "", c.Flags("--debug"), ""); got != true {
		t.Fatalf("expected --debug to parse from os.Args, got %v", got)
	}
}

func TestStringMissingValueExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--name", "--debug"})

	expectExit(t, func() {
		_ = c.GetString("default", "", c.Flags("--name"), "")
	})
}

func TestIntInvalidValueExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--port", "abc"})

	expectExit(t, func() {
		_ = c.GetInt(3000, "", c.Flags("--port"), "")
	})
}

func TestMismatchedAliasValuesExit(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--name=api", "-n", "worker"})

	expectExit(t, func() {
		_ = c.GetString("default", "", c.Flags("--name", "-n"), "")
	})
}

func TestHelpFlagCanBeDisabledExplicitly(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--help=false"})

	if c.IsHelpFlagged() {
		t.Fatal("expected --help=false not to request help")
	}
}

func TestBareHelpFlagRequestsHelp(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--help"})

	if !c.IsHelpFlagged() {
		t.Fatal("expected bare --help to request help")
	}
}

func TestHelpEnvRequestsHelp(t *testing.T) {
	t.Setenv("vibe_help", "yes")

	c := NewCmdParserFromArgs([]string{})

	if !c.IsHelpFlagged() {
		t.Fatal("expected vibe_help=yes to request help")
	}
}
