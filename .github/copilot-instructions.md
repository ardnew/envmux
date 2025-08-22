# Static Environment Generator

This project contains a Go module and command-line tool to generate static environments
from composable, namespaced variables defined using complex expressions
in a custom domain-specific language (DSL).

## Project Structure

The source code is organized into several packages described below.
The descriptions apply to the named package itself and its sub-packages.

- [cmd/envmux](/cmd/envmux): Contains the command-line interface (CLI) implementation.
- [spec](/spec): Handles configuration parsing and management.
- [spec/env](/spec/env): Constructs the environments from a fully-parsed abstract syntax tree (AST).
  - This package is responsible for evaluating expressions used to define variables.
  - Expressions are evaluated using a separate DSL provided by [**_expr-lang/expr_**](https://github.com/expr-lang/expr).
- [spec/parse](/spec/parse): Defines the lexer and parser for the DSL and produces the AST.
  - The lexer and parser are based on [**_pointlander/peg_**](https://github.com/pointlander/peg).
- [spec/parse/internal/grammar.peg](/spec/parse/internal/grammar.peg): Defines the grammar for the DSL using PEG syntax.
- [pkg](/pkg): Provides shared utilities and data structures used across the application.
