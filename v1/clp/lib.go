package clp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (c *CmdParser) ParseBool(str string) bool {
	switch str {
	case "1", "true", "TRUE", "True":
		return true
	case "0", "false", "FALSE", "False":
		return false
	}
	return false
}

type FlagMetaHelp struct {
	Type         string
	EnvName      string
	CommandFlags []string
	DefaultValue interface{}
	Description  string
}

func (c *CmdParser) ParseBoolOptimistic(str string) bool {
	switch str {
	case "0", "false", "FALSE", "False":
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
	NotEnoughArgs bool
	Values        []string
}

func NewCmdParser() *CmdParser {

	m := make(map[string]string)
	metaFlag := make(map[string]MetaFlag)
	nonFlagArgs := []string{}

	var args = os.Args[:]
	var lnArgs = len(os.Args)

	for i, o := range args {

		if !strings.HasPrefix(o, "-") {
			nonFlagArgs = append(nonFlagArgs, o)
			continue
		}

		var metaFlgValue = MetaFlag{
			HasEquals:     false,
			NotEnoughArgs: false,
			Values:        make([]string, 0),
		}

		flg, value := func(o string, i int) (string, string) {
			parts := strings.SplitN(o, "=", 2)
			if len(parts) == 1 {
				if lnArgs <= i+1 {
					// not enough arguments, appending a dummy -- value
					metaFlgValue.NotEnoughArgs = true
					return parts[0], ""
				}
				return parts[0], os.Args[i+1]
			}
			metaFlgValue.HasEquals = true
			return parts[0], parts[1]
		}(o, i)

		m[flg] = value
		metaFlgValue.Values = append(metaFlgValue.Values, value)
		metaFlag[flg] = metaFlgValue
	}

	return &CmdParser{
		FlagsMap:     m,
		FlagsMetaMap: metaFlag,
	}
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
			Stdout.Warning("could not parse int from env var:", env)
		} else {
			ret = z
		}
	}

	var isAlreadySet = false
	for _, v := range flags {

		if v == "" {
			continue
		}

		var value = c.FlagsMap[v]

		if value == "" {
			continue
		}

		parsed, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			Stdout.Warning("could not parse int from command line flag:", v)
			os.Exit(1)
		}

		if isAlreadySet && ret != parsed {
			Stdout.Warning("command line flags are mismatched:", flags)
			Stdout.Warning("command line args were:", os.Args)
			os.Exit(1)
		}

		ret = parsed
		isAlreadySet = true

	}

	return ret
}

func (c *CmdParser) IsHelpFlagged() bool {

	if v, ok := c.FlagsMap["--help"]; ok {
		if c.ParseBoolOptimistic(v) {
			return true
		}
	}

	if c.ParseBool(os.Getenv("vibe_help")) {
		return true
	}

	return false
}

func (c *CmdParser) GetBool(_default bool, env string, flags []string, desc string) bool {

	if c.IsHelpFlagged() {

		c.FlagsHelp = append(c.FlagsHelp, FlagMetaHelp{
			Type:         "bool",
			EnvName:      env,
			CommandFlags: flags,
			DefaultValue: _default,
			Description:  desc,
		})

		Stdout.Info(map[string]interface{}{
			"envVarName":   env,
			"type":         "bool",
			"flags":        flags,
			"defaultValue": _default,
			"description":  desc,
		})
		return false
	}

	ret := _default

	if os.Getenv(env) != "" {
		ret = c.ParseBool(os.Getenv(env))
	}

	var isAlreadySet = false
	for _, v := range flags {

		if v == "" {
			Stdout.Warning("Empty flag:", flags)
			continue
		}

		var value, ok1 = c.FlagsMap[v]
		var metaValue, ok2 = c.FlagsMetaMap[v]

		if !ok1 {
			if ok2 {
				Stdout.Warning("flag was in 1st map but not 2nd map, library error.")
				os.Exit(1)
			}
		}

		if len(metaValue.Values) > 1 {
			Stdout.WarningF("More than one boolean flag at command line: '%v'", v)
		}

		var parsed = true

		if metaValue.HasEquals {
			// boolean can only be false if  -v=false or --v=0, etc
			parsed = c.ParseBoolOptimistic(value)
		}

		if isAlreadySet && ret != parsed {
			Stdout.Warning("command line flags are mismatched:", flags)
			Stdout.Warning("command line args were:", os.Args)
			os.Exit(1)
		}
		ret = parsed
		isAlreadySet = true

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
			Stdout.Warning("Flag is an empty string:", flags)
			continue
		}

		var value = c.FlagsMap[v]

		if value == "" {
			continue
		}

		if isAlreadySet && ret != value {
			Stdout.Warning("command line flags are mismatched:", flags)
			Stdout.Warning("command line args were:", os.Args)
			os.Exit(1)
		}

		ret = value
		isAlreadySet = true

	}

	return ret
}

func (c *CmdParser) Flags(args ...string) []string {
	return args
}
