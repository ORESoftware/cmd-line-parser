package clp

import (
	"strconv"
	"strings"
	"testing"
)

func FuzzNumericDashTokensRemainValues(f *testing.F) {
	for _, seed := range []string{"-1", "-1.5", "-.5", "-1e3", "-0", "-0.0", "-3.14159"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if !strings.HasPrefix(value, "-") || value == "-" {
			return
		}
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return
		}
		if isFlagToken(value) {
			t.Fatalf("numeric token %q was classified as a flag", value)
		}
	})
}

func FuzzEqualsFlagPreservesEntireValue(f *testing.F) {
	for _, seed := range []string{"api", "", "a=b", "--literal", "-1", "hello world"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if strings.ContainsAny(value, "\x00\r\n") {
			return
		}
		flag, occurrence := parseFlagOccurrence("--name=" + value)
		if flag != "--name" || !occurrence.HasEquals || !occurrence.HasValue || occurrence.Value != value {
			t.Fatalf("equals parse drift for %q: flag=%q occurrence=%#v", value, flag, occurrence)
		}
	})
}
