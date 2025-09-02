# Static Environment Generator

This project contains a Go module and command-line tool to generate static environments
from composable, namespaced variables defined using complex expressions
in a custom domain-specific language (DSL).

## Project Structure

The source code is organized into several packages described below.
The descriptions apply to the named package itself and its sub-packages.

- [cmd/envmux](/cmd/envmux): Contains the command-line interface (CLI) implementation.
- [manifest](/manifest): Handles configuration parsing and management.
  - Expressions are evaluated using a separate DSL provided by [**_expr-lang/expr_**](https://github.com/expr-lang/expr).
  - This package is responsible for evaluating expressions used to define variables.
- [manifest/builtin](/manifest/builtin): Defines the variables and built-in functions available for use in the manifest.
- [manifest/parse](/manifest/parse): Defines the lexer and parser for the DSL and produces the AST.
  - The lexer and parser are based on [**_pointlander/peg_**](https://github.com/pointlander/peg).
- [manifest/parse/internal/grammar.peg](/manifest/parse/internal/grammar.peg): Defines the grammar for the DSL using PEG syntax.
- [pkg](/pkg): Provides shared utilities and data structures used across the application.
