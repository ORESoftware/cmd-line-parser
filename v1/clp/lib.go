package clp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseBoolExplicit(str string) (bool, bool) {
	switch strings.ToUpper(strings.TrimSpace(str)) {
	case "1", "TRUE", "T", "YES", "Y", "YASS", "ON":
		return true, true
	case "0", "FALSE", "F", "NO", "N", "OFF":
		return false, true
	default:
		return false, false
	}
}

func (c *CmdParser) ParseBool(str string) bool {
	value, ok := parseBoolExplicit(str)
	return ok && value
}

type FlagMetaHelp struct {
	Type         string
	EnvName      string
	CommandFlags []string
	DefaultValue interface{}
	Description  string
}

// ParseBoolOptimistic is retained for compatibility with callers that
// intentionally want unknown non-empty values to mean true. Command-line
// admission uses the closed parseBoolExplicit vocabulary instead.
func (c *CmdParser) ParseBoolOptimistic(str string) bool {
	switch strings.ToUpper(strings.TrimSpace(str)) {
	case "0", "FALSE", "F", "NO", "N", "OFF":
		return false
	}
	return true
}

type CmdParser struct {
	FlagsMap     map[string]string
	FlagsMetaMap map[string]MetaFlag
	NonFlagArgs  []string
	FlagsHelp    []FlagMetaHelp
}

type MetaFlag struct {
	HasEquals     bool
	HasValue      bool
	NotEnoughArgs bool
	Values        []string
	Occurrences   []FlagOccurrence
}

type FlagOccurrence struct {
	HasEquals     bool
	HasValue      bool
	NotEnoughArgs bool
	Value         string
}

func NewCmdParser() *CmdParser {
	return NewCmdParserFromArgs(os.Args[1:])
}

func NewCmdParserFromArgs(args []string) *CmdParser {
	m := make(map[string]string)
	metaFlag := make(map[string]MetaFlag)
	nonFlagArgs := []string{}

	for i := 0; i < len(args); i++ {
		o := args[i]

		if o == "--" {
			nonFlagArgs = append(nonFlagArgs, args[i+1:]...)
			break
		}

		if !isFlagToken(o) {
			nonFlagArgs = append(nonFlagArgs, o)
			continue
		}

		flg, occurrence := parseFlagOccurrence(o)

		if !occurrence.HasEquals {
			if i+1 < len(args) && shouldConsumeNextValue(args[i+1]) {
				occurrence.Value = args[i+1]
				occurrence.HasValue = true
				i++
			} else {
				occurrence.NotEnoughArgs = true
			}
		}

		m[flg] = occurrence.Value

		metaFlgValue := metaFlag[flg]
		metaFlgValue.HasEquals = occurrence.HasEquals
		metaFlgValue.HasValue = occurrence.HasValue
		metaFlgValue.NotEnoughArgs = occurrence.NotEnoughArgs
		metaFlgValue.Values = append(metaFlgValue.Values, occurrence.Value)
		metaFlgValue.Occurrences = append(metaFlgValue.Occurrences, occurrence)
		metaFlag[flg] = metaFlgValue
	}

	return &CmdParser{
		FlagsMap:     m,
		FlagsMetaMap: metaFlag,
		NonFlagArgs:  nonFlagArgs,
	}
}

func isFlagToken(value string) bool {
	if value == "" || value == "-" {
		return false
	}

	if !strings.HasPrefix(value, "-") {
		return false
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return false
	}

	return true
}

func shouldConsumeNextValue(value string) bool {
	return value != "--" && !isFlagToken(value)
}

func parseFlagOccurrence(arg string) (string, FlagOccurrence) {
	parts := strings.SplitN(arg, "=", 2)
	occurrence := FlagOccurrence{}

	if len(parts) == 1 {
		return parts[0], occurrence
	}

	occurrence.HasEquals = true
	occurrence.HasValue = true
	occurrence.Value = parts[1]
	return parts[0], occurrence
}

func occurrencesForMeta(meta MetaFlag) []FlagOccurrence {
	if len(meta.Occurrences) > 0 {
		return meta.Occurrences
	}

	occurrences := make([]FlagOccurrence, 0, len(meta.Values))
	for _, value := range meta.Values {
		occurrences = append(occurrences, FlagOccurrence{
			HasEquals:     meta.HasEquals,
			HasValue:      meta.HasValue || meta.HasEquals || value != "",
			NotEnoughArgs: meta.NotEnoughArgs,
			Value:         value,
		})
	}

	return occurrences
}

var exitProcess = os.Exit

func exitWithFlagMismatch(flags []string) {
	Stdout.Warn("command line flags are mismatched:", flags)
	Stdout.Warn("command line args were:", os.Args)
	exitProcess(1)
}

func exitWithMissingFlagValue(flag string, valueType string) {
	Stdout.Warn("missing", valueType, "value for command line flag:", flag)
	Stdout.Warn("command line args were:", os.Args)
	exitProcess(1)
}

func exitWithInvalidFlagValue(flag string, valueType string, value string) {
	Stdout.Warn("invalid", valueType, "value for command line flag:", flag, value)
	Stdout.Warn("command line args were:", os.Args)
	exitProcess(1)
}

func (c *CmdParser) GetInt(_default int64, env string, flags []string, desc string) int64 {

	if c.IsHelpFlagged() {

		c.FlagsHelp = append(c.FlagsHelp, FlagMetaHelp{
			Type:         "int",
			EnvName:      env,
			CommandFlags: flags,
			DefaultValue: _default,
			Description:  desc,
		})

		Stdout.Info(map[string]interface{}{
			"envVar":       env,
			"type":         "int",
			"flags":        flags,
			"defaultValue": _default,
			"description":  desc,
		})
		return 0
	}

	ret := _default

	if os.Getenv(env) != "" {
		if z, err := strconv.ParseInt(os.Getenv(env), 10, 64); err != nil {
			Stdout.Warn("could not parse int from env var:", env)
		} else {
			ret = z
		}
	}

	var isAlreadySet = false
	for _, v := range flags {

		if v == "" {
			continue
		}

		metaValue, ok := c.FlagsMetaMap[v]
		if !ok {
			continue
		}

		if len(metaValue.Values) > 1 {
			Stdout.WarnF("More than one int flag at command line: '%v'", v)
		}

		for _, occurrence := range occurrencesForMeta(metaValue) {
			if !occurrence.HasValue {
				exitWithMissingFlagValue(v, "int")
			}

			parsed, err := strconv.ParseInt(occurrence.Value, 10, 64)

			if err != nil {
				Stdout.Warn("could not parse int from command line flag:", v)
				exitProcess(1)
			}

			if isAlreadySet && ret != parsed {
				exitWithFlagMismatch(flags)
			}

			ret = parsed
			isAlreadySet = true
		}
	}

	return ret
}

func (c *CmdParser) IsHelpFlagged() bool {

	if meta, ok := c.FlagsMetaMap["--help"]; ok {
		occurrences := occurrencesForMeta(meta)
		if len(occurrences) == 0 {
			return true
		}

		last := occurrences[len(occurrences)-1]
		if !last.HasValue {
			return true
		}

		parsed, valid := parseBoolExplicit(last.Value)
		if !valid {
			exitWithInvalidFlagValue("--help", "boolean", last.Value)
			return false
		}
		return parsed
	}

	if c.ParseBool(os.Getenv("vibe_help")) {
		return true
	}

	return false
}

func (c *CmdParser) GetBool(defaultValue bool, env string, flags []string, desc string) bool {

	if c.IsHelpFlagged() {

		c.FlagsHelp = append(c.FlagsHelp, FlagMetaHelp{
			Type:         "bool",
			EnvName:      env,
			CommandFlags: flags,
			DefaultValue: defaultValue,
			Description:  desc,
		})

		Stdout.Info(map[string]interface{}{
			"envVarName":   env,
			"type":         "bool",
			"flags":        flags,
			"defaultValue": defaultValue,
			"description":  desc,
		})
		return false
	}

	ret := defaultValue

	if os.Getenv(env) != "" {
		ret = c.ParseBool(os.Getenv(env))
	}

	var isAlreadySet = false
	for _, v := range flags {

		if v == "" {
			Stdout.Warn("Empty flag:", flags)
			continue
		}

		metaValue, ok := c.FlagsMetaMap[v]
		if !ok {
			continue
		}

		if len(metaValue.Values) > 1 {
			Stdout.WarnF("More than one boolean flag at command line: '%v'", v)
		}

		for _, occurrence := range occurrencesForMeta(metaValue) {
			parsed := true

			if occurrence.HasValue {
				var valid bool
				parsed, valid = parseBoolExplicit(occurrence.Value)
				if !valid {
					exitWithInvalidFlagValue(v, "boolean", occurrence.Value)
					continue
				}
			}

			if isAlreadySet && ret != parsed {
				exitWithFlagMismatch(flags)
			}

			ret = parsed
			isAlreadySet = true
		}
	}

	return ret
}

func (c *CmdParser) PrintHelp() {
	Stdout.Info("Help / command line args/env:")
	fmt.Println("")
	fmt.Println("Here are the env vars and command line flags:")
	fmt.Println("")
	for _, v := range c.FlagsHelp {
		fmt.Println("\t", "Env / flags:", v.EnvName, v.CommandFlags)
		fmt.Println("\t\t", fmt.Sprintf("Type: '%v'", v.Type))
		fmt.Println("\t\t", fmt.Sprintf("Default value: '%v'", v.DefaultValue))
		fmt.Println("\t\t", fmt.Sprintf("Description: '%v'", v.Description))
		fmt.Println("")
	}
}

func (c *CmdParser) GetString(_default string, env string, flags []string, desc string) string {

	if c.IsHelpFlagged() {

		c.FlagsHelp = append(c.FlagsHelp, FlagMetaHelp{
			Type:         "string",
			EnvName:      env,
			CommandFlags: flags,
			DefaultValue: _default,
			Description:  desc,
		})

		Stdout.Info(map[string]interface{}{
			"envVarName":   env,
			"type":         "string",
			"flags":        flags,
			"defaultValue": _default,
			"description":  desc,
		})
		return ""
	}

	ret := _default

	if os.Getenv(env) != "" {
		ret = os.Getenv(env)
	}

	var isAlreadySet = false
	for _, v := range flags {

		if v == "" {
			Stdout.Warn("Flag is an empty string:", flags)
			continue
		}

		metaValue, ok := c.FlagsMetaMap[v]
		if !ok {
			continue
		}

		if len(metaValue.Values) > 1 {
			Stdout.WarnF("More than one string flag at command line: '%v'", v)
		}

		for _, occurrence := range occurrencesForMeta(metaValue) {
			if !occurrence.HasValue {
				exitWithMissingFlagValue(v, "string")
			}

			if isAlreadySet && ret != occurrence.Value {
				exitWithFlagMismatch(flags)
			}

			ret = occurrence.Value
			isAlreadySet = true
		}
	}

	return ret
}

func (c *CmdParser) Flags(args ...string) []string {
	return args
}
