# envmux

> Static environment compositor

[![Go Reference](https://pkg.go.dev/badge/github.com/envmux/envmux.svg)](https://pkg.go.dev/github.com/envmux/envmux)
[![Go Report Card](https://goreportcard.com/badge/github.com/envmux/envmux)](https://goreportcard.com/report/github.com/envmux/envmux)

## _What is this I don't even_

`envmux` evaluates complex expressions assigned to namespaced variables defined in manifest files.

Manifests are simple, readable, and composable. Namespaces group related variables, can include other namespaces, and can be parameterized. Expressions are evaluated with a safe expression engine.

--

Minimal manifest example:

```text
# file: example.env
base {
  APP_NAME = "envmux";
  USERNAME = user.Username;     # builtin 'user' from host system
  GREETING = "Hello, " + USERNAME;
}
```

Namespaces can be composed of other namespaces by definition.

Composition example:

```text
base {
  FOO = "bar";
}

app <base> {                    # app imports all variables from base
  MSG = FOO + "!";            # uses FOO defined in base
}
```

Namespaces can also be parameterized either by definition or by composition.

Parametric namespaces (access parameter via '_' in expressions):

```text
# Evaluate once per parameter value and merge results
greet ("world", "team") {
  HELLO = "Hello, " + _;      # _ is the implicit parameter
}

# Pass parameters during composition
banner <greet("ops")> {
  SHOUT = HELLO + "!!!";      # uses HELLO from composed greet
}
```

### Key Features

- **Namespace Composition**: Construct process environments independent of the shell
- **Efficient Parsing**: Manifest parsed using a PEG-backed grammar ([pointlander/peg](https://github.com/pointlander/peg))
- **Expression Support**: Define variables using a rich expression language ([expr-lang/expr](https://github.com/expr-lang/expr))
- **Flexible Input**: Read manifests from files, stdin, or directly from command-line arguments
- **Parallel Processing**: Evaluate environments efficiently with configurable parallelism
- **Subcommands**: Manage file system operations and namespace operations

## Installation

```sh
go install github.com/envmux/envmux@latest
```

## Usage

```sh
# Load default manifest and output variables from the 'development' namespace
envmux development

# Use a specific manifest file
envmux -m path/to/manifest.env development

# Define inline namespaces
envmux -d 'test { FOO = "bar"; }' test

# Compose multiple namespaces
envmux dev base

# Use with other commands
eval $(envmux production)
```

### Command-Line Options

```sh
Usage: envmux [flags] [subcommand ...]

Flags:
  -V, --version             Show semantic version
  -v, --verbose             Enable verbose output
  -i, --ignore-default      Ignore default manifest file
  -s, --strict-definitions  Treat undefined namespaces as errors
  -j, --jobs N              Maximum number of parallel tasks (default: CPU cores)
  -c, --config FILE         Config file with default flags
  -m, --manifest FILE       Manifest file containing namespace definitions ("-" is stdin)
  -d, --define SOURCE       Inline namespace definitions to append
```

If compiled with `-tags pprof`, an additional profiling flag is available:

```sh
  -p, --profile TYPE        enable profiling (mem|allocs|heap|mutex|thread|trace|block|cpu|goroutine|clock)
```

## License

This project is licensed under the terms found in the [LICENSE](LICENSE) file.
