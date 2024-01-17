package clp

import (
	json_logging "github.com/oresoftware/json-logging/jlog/lib"
	ll "github.com/oresoftware/json-logging/jlog/level"
	"os"
)

var appName = os.Getenv("app_name")

var Stdout = json_logging.CreateLogger(appName).SetLogLevel(ll.TRACE).SetOutputFile(os.Stdout)
var Stderr = json_logging.CreateLogger(appName).SetLogLevel(ll.WARN).SetOutputFile(os.Stderr)
