package clp

import (
	json_logging "github.com/oresoftware/json-logging/jlog/lib"
	"github.com/oresoftware/json-logging/jlog/shared"
	"os"
)

var appName = os.Getenv("app_name")

var Stdout = json_logging.CreateLogger(appName).SetLogLevel(shared.TRACE)
var Stderr = json_logging.CreateLogger(appName).SetLogLevel(shared.WARN)
