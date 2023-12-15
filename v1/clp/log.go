package clp

import (
	json_logging "github.com/oresoftware/json-logging/jlog"
	"os"
)

var appName = os.Getenv("app_name")

var Stdout = json_logging.New(appName, false, "hostname here")
var Stderr = json_logging.New(appName, false, "hostname here")
