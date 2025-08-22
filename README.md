# envmux — Static Environment Generator

Compose, evaluate, and export reproducible, namespaced environments from a concise DSL with expression evaluation, composition, and parallel execution.

> Generate static, declarative environments for CI/CD, build pipelines, and local workflows.

---

## Contents

1. [envmux — Static Environment Generator](#envmux--static-environment-generator)
   1. [Contents](#contents)
   2. [Features](#features)
   3. [Quick Start](#quick-start)
   4. [Installation](#installation)
   5. [DSL Overview](#dsl-overview)
   6. [Examples](#examples)
      1. [Basic Inheritance](#basic-inheritance)
      2. [Path Composition](#path-composition)
      3. [Dynamic Expression](#dynamic-expression)
      4. [Multi-Parent Merge (later overrides earlier)](#multi-parent-merge-later-overrides-earlier)
      5. [Inline Override (CLI)](#inline-override-cli)
   7. [CLI Usage](#cli-usage)
   8. [Inline Definitions](#inline-definitions)
   9. [Configuration \& Defaults](#configuration--defaults)
   10. [Architecture](#architecture)
   11. [Evaluation Model](#evaluation-model)
   12. [Error Handling](#error-handling)
   13. [Profiling \& Performance](#profiling--performance)
   14. [Testing \& Development](#testing--development)
   15. [FAQ](#faq)

---

## Features

- **Namespaced Environments** with inheritance and composition.
- **Declarative DSL** backed by a PEG grammar ([`spec/parse/internal/grammar.peg`](spec/parse/internal/grammar.peg)).
- **Expression Evaluation** via [`expr-lang/expr`](https://github.com/expr-lang/expr) (safe, extensible).
- **Parallel Evaluation** of namespaces (configurable worker limit).
- **Deterministic Output**: process environment is not implicitly inherited.
- **Inline Definitions**: augment manifests directly from the CLI.
- **Rich Built-ins**: path ops, file helpers, platform metadata, user info, coercion utilities.
- **Extensible Packages**:
  - Functional helpers: [`pkg/fn`](pkg/fn)
  - Environment var normalization: [`pkg/ev`](pkg/ev)
  - Error taxonomy: [`pkg/errs`](pkg/errs)
  - Runtime defaults & discovery: [`pkg/run`](pkg/run)
  - Profiling hooks: [`pkg/prof`](pkg/prof)
- **Clear Separation** between parsing (`spec/parse`) and evaluation (`spec/env`).
- **Zero-Config Defaults**: automatic default manifest & config file locations.
- **Machine-Friendly Output**: standard `KEY=VALUE` lines.

---

## Quick Start

```sh
go install github.com/ardnew/envmux/cmd/envmux@latest
```

Create `env.def`:

```text
base {
  ROOT = "/opt/tool";
  BIN  = path.cat(ROOT, "bin");
}

dev <base> {
  DEBUG = true;
  PATH  = path.cat(BIN, "/usr/bin");
}
```

Evaluate:

```sh
envmux -m env.def dev
```

Output:

```
ROOT=/opt/tool
BIN=/opt/tool/bin
DEBUG=true
PATH=/opt/tool/bin:/usr/bin
```

---

## Installation

| Method | Command |
| ------ | ------- |
| Latest tagged release | `go install github.com/ardnew/envmux/cmd/envmux@latest` |
| Specific version | `go install github.com/ardnew/envmux/cmd/envmux@vX.Y.Z` |
| From source | `git clone https://github.com/ardnew/envmux.git && cd envmux && go build -o envmux ./cmd/envmux` |

Regenerate parser after grammar changes:

```sh
go generate ./...
```

> [!NOTE]
> Go 1.24+ is required (`go.mod` sets `go 1.24.5`).

---

## DSL Overview

Core constructs:

| Concept | Description |
| ------- | ----------- |
| Namespace | Named block containing variable assignments & parameters |
| Composition | `child <parent other>` merges definitions (left-to-right) |
| Assignment | `NAME = <expr>` |
| Expressions | Backed by `expr-lang/expr`; support arithmetic, logic, ternaries, pipelines |
| Literals | Strings `'x'` / `"x"`, numbers, booleans, `nil` |
| Comments | `//`, `#`, `/* ... */` |

Minimal grammar sketch (see the full PEG in [`spec/parse/internal/grammar.peg`](spec/parse/internal/grammar.peg)):

```
Namespace  <- Ident (WS+ Ident)* (WS* '<' WS* Parents WS*)? WS* '{' Body '}'
Body       <- (Stmt (';' / NL)*)*
Stmt       <- Assign
Assign     <- Ident WS* '=' WS* Expr
Expr       <- pipeline / logical / arithmetic / primary ...
```

> [!TIP]
> Use pipelines: `FOO = "hello" | upper() | trim()`.

---

## Examples

### Basic Inheritance

```text
default {
  APP = "demo";
  DEBUG = false;
}

dev <default> {
  DEBUG = true;
  LOG_LEVEL = "debug";
}
```

### Path Composition

```text
paths {
  ROOT = "/opt/project";
  BIN  = path.cat(ROOT, "bin");
  PATH = path.cat(BIN, "/usr/bin");
}
```

### Dynamic Expression

```text
calc {
  BASE = 10;
  TWICE = BASE * 2;
  MAYBE = user.Name ?? "anonymous";
}
```

### Multi-Parent Merge (later overrides earlier)

```text
prod <default paths> {
  DEBUG = false;
}
```

### Inline Override (CLI)

```sh
envmux -d 'hotfix <prod> { DEBUG = true; }' hotfix
```

---

## CLI Usage

```
envmux [flags] [namespace ...]
```

Common flags:

| Flag | Description |
| ---- | ----------- |
| `-m, --manifest FILE` | Manifest file (repeatable, `-` = stdin) |
| `-d, --define SOURCE` | Inline namespace block appended after manifests |
| `-j, --jobs N` | Max parallel evaluation workers |
| `-u, --require-definitions` | Error on unknown namespaces |
| `-i, --ignore-default` | Do not load default manifest |
| `-c, --config FILE` | Config file of default flags |
| `-v, --verbose` | Increase verbosity (repeat for higher levels) |
| `-V, --version` | Show version |
| `-b, --buffer-size N` | Parser buffer size (bytes or SI units) |
| `-p, --profile TYPE` | Enable profiling (build with `-tags pprof`) |

If no namespaces are provided, a fallback (e.g. current working directory name) may be used—see [`pkg/run`](pkg/run).

---

## Inline Definitions

Order of application:

1. All manifest files (in provided order, plus default unless suppressed)
2. All `--define` sources (in provided order)
3. Namespace selection & evaluation

This allows targeted overrides without editing files.

---

## Configuration & Defaults

- Default config directory: resolved via [`pkg/run`](pkg/run) (platform-aware).
- Default manifest autoloaded unless `--ignore-default`.
- Config file (flag `--config`) may contain flags (format per [`github.com/peterbourgon/ff/v4`](https://github.com/peterbourgon/ff)).

---

## Architecture

| Layer | Package | Summary |
| ----- | ------- | ------- |
| CLI | [`cmd/envmux`](cmd/envmux) | Entry point, command graph, flag wiring |
| Parse | [`spec/parse`](spec/parse) | PEG parser, AST build, buffer management |
| Environment Model | [`spec/env`](spec/env) | Namespace model, composition, evaluation pipeline |
| Utilities | [`pkg/fn`](pkg/fn) | Functional helpers (Option, Map, Filter, Unique) |
| Environment Utilities | [`pkg/ev`](pkg/ev) | Env var normalization & formatting |
| Errors | [`pkg/errs`](pkg/errs) | Structured error types (`IncompleteParseError`, etc.) |
| Runtime Defaults | [`pkg/run`](pkg/run) | Config dir, namespace derivation |
| Profiling | [`pkg/prof`](pkg/prof) | Conditional pprof initialization |

Key symbol (evaluation entry):
[`env.Model`](spec/env) — constructed via [`env.Make`](spec/env), then chained:

```go
mod, err := env.Make(ctx, manifests, inline,
  env.WithMaxParallelJobs(N),
  env.WithEvalRequiresDef(require),
)
mod, err = mod.Parse()
out, err := mod.Eval(ctx, "namespace")
```

---

## Evaluation Model

1. **Parse Phase**: Build AST from manifests + inline sources (buffer sized via `--buffer-size`).
2. **Normalization**: Validate identifiers, composition graph.
3. **Dependency Planning**: Determine evaluation order, detect cycles.
4. **Parallel Execution**: Evaluate independent namespaces concurrently (bounded by `--jobs`).
5. **Expression Resolution**: Each assignment executed in an `expr` environment seeded with:
   - Namespace variables (respecting override order)
   - Built-in funcs (path ops, file system queries, string utilities, numeric helpers)
   - Host metadata (user, platform) exposed via structured values
6. **Export**: Produce stable `[]string` of `KEY=VALUE`.

> [!IMPORTANT]
> Process environment variables are not implicitly imported—declare what you need explicitly for reproducibility.

---

## Error Handling

Representative error types (see [`pkg/errs`](pkg/errs)):

| Error | Cause |
| ----- | ----- |
| `errs.IncompleteParseError` | Parser accepted only part of input |
| `errs.ErrInvalidDefinitions` | Malformed or conflicting manifest inputs |
| `errs.ErrIncompleteEval` | Runtime evaluation error (cycle, missing dependency) |
| `errs.ExpressionError` | Expression compilation or execution failure |

Wraps preserve context; callers can use `errors.Is` / `errors.As`.

---

## Profiling & Performance

Build with profiling:

```sh
go build -tags pprof ./cmd/envmux
./envmux -p cpu -p mem -m env.def target
```

Available modes listed via `-p help` or `--profile` usage (see [`pkg/prof`](pkg/prof)).

Performance guidelines:

- Keep expressions side-effect free.
- Use composition over repetition.
- Avoid large intermediate strings; prefer concatenation through provided helpers.
- Tune `--jobs` for I/O vs CPU balance.
- Adjust `--buffer-size` for very large manifests to reduce reallocations.

---

## Testing & Development

Run tests:

```sh
go test ./...
```

Add parser changes → regenerate:

```sh
go generate ./...
go test ./spec/...
```

Focused packages with tests:

- Functional utilities: [`pkg/fn`](pkg/fn)
- Env var normalization: [`pkg/ev`](pkg/ev)

Suggested workflow:

1. Modify grammar (`spec/parse/internal/grammar.peg`)
2. `go generate ./spec/parse/...`
3. Add / adjust tests
4. Benchmark critical transformations if changed
5. Commit with clear semantic message (e.g. `feat(parse): ...`)

---

## FAQ

**Q: Why not load my shell environment automatically?**
A: Determinism—explicit definitions ensure reproducible builds.

**Q: How do I override a single variable for a namespace?**
Use an inline definition with composition:

```sh
envmux -d 'tmp <prod> { FEATURE_X = false; }' tmp
```

**Q: Can I reference another variable before it is defined?**
Yes, if it resides in a composed parent or earlier in the resolution order; cycles are detected and reported.

**Q: How do I inspect the parse step only?**
Run with a nonexistent namespace and `--require-definitions` to surface parse diagnostics without evaluation.

---

> [!TIP]
> Explore the manifest examples and internal docs in [`docs`](docs) and the grammar reference at [`spec/parse/internal/grammar.peg`](spec/parse/internal/grammar.peg)
