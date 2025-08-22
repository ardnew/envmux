package docs

// Grammar for the envmux manifest file containing namespace definitions.
// See: https://github.com/ardnew/envmux/blob/main/docs/manifest.md
//
// The envmux manifest file defines one or more namespaces, each of which
// can contain parameters, variable definitions, and can be composed of other
// namespaces.
//
// Variables are evaluated as [github.com/expr-lang/expr] expressions.
// The evaluation environment is derived from the enclosing namespace along
// with many built-in functions and runtime variables inherited from the host
// system. Note that the process environment is not inherited by default.
//
//		// A simple namespace "foo" with two variables
//		foo {
//		 	bar = "HELLO";        // A simple string variable
//		 	baz = bar | lower();  // baz = "hello"
//		}
//
//	 // A namespace "plugh xyzzy" composed of "foo"
//		plugh xyzzy <foo> {
//		 	quux = baz + ", " + user.Name; // quux = "hello, <USERNAME>"
//		}
//
// The grammar is divided into the following sections:
//
//	 - Whitespace and Comments
//	 - Identifiers
//	 - Literals
//	 - Expressions
//	 - Statements
//	 - Namespaces
//
// ## Whitespace and Comments
//
// Whitespace is defined as one or more spaces, tabs, or newlines. It is
// ignored by the parser and can be used to format the manifest file for
// readability.
//
//	/* This is a block comment*/
//	foo {
//	 	// Comments can appear anywhere
//	 	bar = "baz"; // Line comment
//	}
//
// ## Identifiers
//
// Identifiers are used to name namespaces, variables, and parameters.
// They must start with a letter or underscore, followed by letters, digits, or
// underscores.
//
//		// Namespaces must start with an identifier, but can be followed by any
//	 // number of alphanumeric characters and whitespace (excluding newlines).
//
//		// Valid identifiers
//		foo bar {
//		 	bar_baz = "hello"
//		 	qux = "world"
//		}
//
// ## Literals
//
// Literals are constant values that are represented directly in the manifest
// file. The following literal types are supported:
//
//	 - Strings: `"hello"` or `'hello'`
//	 - Numbers: `123` or `-123.45`
//	 - Booleans: `true` or `false`
//	 - Nil: `nil`
//
// ## Expressions
//
// Expressions are used to compute values based on variables, parameters, and
// literals. They are written in a syntax similar to JavaScript or Python and
// can be as simple or complex as needed.
//
//	// Simple expressions
//	foo = 1 + 2
//	bar = "hello" + " " + "world"
//
//	// Complex expressions
//	baz = user.Name ?? (1 + 2 * 3)
//
// Expressions are evaluated at runtime, and their values can change based on
// the environment and parameters.
//
// ## Statements
//
// Statements are used to perform actions, such as printing to the console or
// defining variables and parameters.
//
//	// A variable assignment
//	foo = "bar"
//
// ## Namespaces
//
// Namespaces are the top-level containers in an envmux manifest file. Each
// namespace can contain variables, parameters, and other namespaces. Namespaces
// are defined using the `namespace` keyword, followed by the namespace name
// and an optional set of curly braces containing the namespace body.
//
//	// A simple namespace "foo" with two variables
//	foo {
//	 	bar = "HELLO";        // A simple string variable
//	 	baz = bar | lower();  // baz = "hello"
//	}
//
//	// A namespace "plugh xyzzy" composed of "foo"
//	plugh xyzzy <foo> {
//	 	quux = baz + ", " + user.Name; // quux = "hello, <USERNAME>"
//	}
