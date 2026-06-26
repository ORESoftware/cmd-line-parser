<!-- BEGIN k8s-cluster-submodule-notice -->
> [!NOTE]
> **Canonical source.** This repository is the source of truth for its code. It
> is also vendored as a **secondary** git submodule of
> [ORESoftware/k8s-cluster](https://github.com/ORESoftware/k8s-cluster) at
> `remote/modules/github/oresoftware/cmd-line-parser` — make changes here, not in that submodule checkout.
>
> On disk: submodule checkout `~/codes/ores/k8s-cluster/remote/modules/github/oresoftware/cmd-line-parser`.
<!-- END k8s-cluster-submodule-notice -->

# cmd-line-parser

Small Go helper for reading config from defaults, environment variables, and command-line flags.

Precedence is:

1. default value
2. environment variable, when the env var is present and non-empty
3. command-line flag

## Usage

```go
package main

import "github.com/oresoftware/cmd-line-parser/v1/clp"

func main() {
	c := clp.NewCmdParser()

	debug := c.GetBool(false, "app_debug", c.Flags("--debug"), "enable debug logs")
	port := c.GetInt(3000, "app_port", c.Flags("--port", "-p"), "HTTP port")
	host := c.GetString("localhost", "app_host", c.Flags("--host"), "HTTP host")

	_, _, _ = debug, port, host
}
```

## Flag Forms

Supported value forms:

- `--host api.local`
- `--host=api.local`
- `--port 3000`
- `--port=3000`
- `--debug`
- `--debug=false`
- `--debug 0`

Use `--` to stop flag parsing:

```sh
app --debug -- --not-a-flag positional
```

String and int flags must have a value. If a value begins with `-`, pass it with equals
syntax, for example `--host=-internal`. Negative numeric values like `--port -1` are accepted.

Boolean flags without a value are treated as `true`. Boolean false values include `0`,
`false`, `f`, `no`, `n`, and `off`.

Repeated aliases must agree:

```sh
app --port 3000 -p 3000
```

If repeated aliases provide different values, the parser logs a warning and exits with code 1.
