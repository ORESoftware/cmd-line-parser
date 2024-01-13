package clp

import (
	json_logging "github.com/oresoftware/json-logging/jlog"
	"os"
)

var appName = os.Getenv("app_name")

var Stdout = json_logging.New(appName, "", json_logging.TRACE, []*os.File{os.Stdout})
var Stderr = json_logging.New(appName, "", json_logging.WARN, []*os.File{os.Stderr})
