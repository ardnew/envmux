# envmux

A static environment generator that composes namespaced variables using a custom domain-specific language.

## What It Do™

envmux constructs and evaluates environment variables from namespaces defined in manifest files. It uses a custom DSL that supports complex expressions, namespace composition, and parameter passing.

### Key Features

- Namespace Composition: Construct process environments with shell agnosticism
- Efficient Parsing: Manifest parsed using a PEG-backed grammar ([pointlander/peg](https://github.com/pointlander/peg))
- Expression Support: Define variables using a rich expression language ([expr-lang/expr](https://github.com/expr-lang/expr))
- Flexible Input: Read manifests from files, stdin, or directly from command-line arguments
- Parallel Processing: Evaluate environments efficiently with configurable parallelism
- Subcommands: Manage file system operations and namespace operations

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
