package clp

import "testing"

func TestExplicitBooleanVocabulary(t *testing.T) {
	trueValues := []string{"1", "true", "T", "yes", "Y", "yass", "ON"}
	for _, value := range trueValues {
		value := value
		t.Run("true_"+value, func(t *testing.T) {
			c := NewCmdParserFromArgs([]string{"--debug=" + value})
			if got := c.GetBool(false, "", c.Flags("--debug"), ""); !got {
				t.Fatalf("expected %q to parse true", value)
			}
		})
	}

	falseValues := []string{"0", "false", "F", "no", "N", "OFF"}
	for _, value := range falseValues {
		value := value
		t.Run("false_"+value, func(t *testing.T) {
			c := NewCmdParserFromArgs([]string{"--debug=" + value})
			if got := c.GetBool(true, "", c.Flags("--debug"), ""); got {
				t.Fatalf("expected %q to parse false", value)
			}
		})
	}
}

func TestExplicitBooleanTypoExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug=flase"})
	expectExit(t, func() {
		_ = c.GetBool(false, "", c.Flags("--debug"), "")
	})
}

func TestSeparatedBooleanGarbageExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug", "definitely"})
	expectExit(t, func() {
		_ = c.GetBool(false, "", c.Flags("--debug"), "")
	})
}

func TestExplicitEmptyBooleanExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug="})
	expectExit(t, func() {
		_ = c.GetBool(false, "", c.Flags("--debug"), "")
	})
}

func TestInvalidHelpBooleanExits(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--help=maybe"})
	expectExit(t, func() {
		_ = c.IsHelpFlagged()
	})
}

func TestBareBooleanStillMeansTrue(t *testing.T) {
	c := NewCmdParserFromArgs([]string{"--debug"})
	if got := c.GetBool(false, "", c.Flags("--debug"), ""); !got {
		t.Fatal("bare bool flag must remain true")
	}
}
