# Static Environment Generator

This project contains a Go module and command-line tool to generate static environments
from composable, namespaced variables defined using complex expressions
in a custom domain-specific language (DSL).

## Project Structure

The source code is organized into several packages described below.
The descriptions apply to the named package itself and its sub-packages.

- [cmd/envmux](/cmd/envmux): Contains the command-line interface (CLI) implementation.
- [config](/config): Handles configuration parsing and management.
- [config/env](/config/env): Constructs the environments from a fully-parsed abstract syntax tree (AST).
  - This package is responsible for evaluating expressions used to define variables.
  - Expressions are evaluated using a separate DSL provided by [**_expr-lang/expr_**](https://github.com/expr-lang/expr).
- [config/parse](/config/parse): Defines the lexer and parser for the DSL and produces the AST.
  - The lexer and parser are based on [**_alecthomas/participle_**](https://github.com/alecthomas/participle).
- [config/parse/stream](/config/parse/stream): Implements streaming for the lexer and parser using a pipeline of token processing stages.
- [pkg](/pkg): Provides shared utilities and data structures used across the application.
