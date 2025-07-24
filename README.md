# envmux

> **Static Environment Generator**  
> Compose, evaluate, and manage namespaced environments with a powerful DSL.

---

envmux is a command-line tool and Go module for generating static environments from composable, namespaced variables defined using a custom domain-specific language (DSL). It is designed for advanced configuration scenarios, CI/CD pipelines, reproducible builds, and any workflow that benefits from declarative, composable environment management.

---

## Features

- **Namespaced Environments:**  
  Define multiple environments with inheritance and composition.

- **Custom DSL:**  
  Express complex variable logic and relationships using a concise, readable syntax.

- **Expression Evaluation:**  
  Integrates [expr-lang/expr](https://github.com/expr-lang/expr) for powerful, type-safe expressions.

- **Parallel Evaluation:**  
  Evaluate environments in parallel for maximum performance.

- **CLI & Go API:**  
  Use as a standalone CLI or as a Go library in your own projects.

- **Extensible:**  
  Easily add new functions, variables, and integrations.

---

## Quick Start

### Install

```sh
go install github.com/ardnew/envmux/cmd/envmux@latest
```

Or build from source:

```sh
git clone https://github.com/ardnew/envmux.git
cd envmux
go build -o envmux ./cmd/envmux
```

### Example Usage

Create a file called `env.def`:

```text
default {
  FOO = "bar";
  PATH = path.cat("/usr/local/bin", "/usr/bin");
}

dev <default> {
  FOO = "devbar";
  DEBUG = true;
}
```

Generate the environment for the `dev` namespace:

```sh
envmux -s env.def dev
```

**Output:**
```
FOO=devbar
PATH=/usr/local/bin:/usr/bin
DEBUG=true
```

---

## DSL Overview

- **Namespaces:**  
  Group variables under named environments.
- **Inheritance:**  
  Use `<parent>` to inherit from other namespaces.
- **Expressions:**  
  Use arithmetic, string, and custom functions.
- **Comments:**  
  Supports `//`, `#`, and `/* ... */` comments.

```text
prod <default> {
  FOO = "prodbar";
  TIMEOUT = 30 * 2;
}
```

---

## CLI Reference

```sh
envmux [flags] [subcommand ...]
```

### Common Flags

| Flag                | Description                                 |
|---------------------|---------------------------------------------|
| `-s, --source`      | Path to environment definitions file        |
| `-j, --jobs`        | Max parallel evaluation jobs (default: CPU) |
| `-i, --ignore-default` | Ignore default namespace definitions     |
| `-u, --require-definitions` | Error on undefined namespaces      |
| `-v, --verbose`     | Enable verbose output                       |
| `-V, --version`     | Show version                                |

### Subcommands

- `fs` — File system management
- `ns` — Namespace operations

Run `envmux --help` or `envmux <subcommand> --help` for details.

---

## Project Structure

- [`cmd/envmux`](cmd/envmux) — CLI entry point
- [`config`](config) — Configuration parsing and management
  - [`config/env`](config/env) — Environment evaluation
  - [`config/parse`](config/parse) — DSL parser (legacy)
  - [`config/parse2`](config/parse2) — Next-gen parser (WIP)
- [`pkg`](pkg) — Shared utilities

---

## Advanced Topics

> [!TIP]
> See the [doc/internal/grammar.html](doc/internal/grammar/internal/grammar.html) for a visual reference of the DSL grammar.

- **Expression Language:**  
  Leverages [expr-lang/expr](https://github.com/expr-lang/expr) for advanced logic.
- **Custom Functions:**  
  Use built-in helpers like `path.cat`, `file.exists`, `cwd()`, etc.
- **Parallelism:**  
  Control evaluation concurrency with `--jobs`.

---

## Development

- Go 1.24+ required
- Run `go generate ./...` to regenerate the lexer/parser after grammar changes.
- Linting and formatting via `golangci-lint` and `gofmt`.

---

> [!NOTE]
> For troubleshooting, usage examples, and more, see the [doc](doc/) directory and inline CLI