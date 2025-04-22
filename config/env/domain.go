package env

type (
	// Lineage is a chain of namespaces that produce a domain.
	lineage []string

	// Plan is a map of environment variables to expressions.
	//
	// It is composed of all [parse.Mapping] elements
	// from each of the namespaces in the [lineage] of its [domain].
	//
	// Each expression is evaluated after its [domain] is fully constructed.
	//
	// A [Context] is required to evaluate expressions,
	// and is passed along with its subject to the expression evaluator.
	//
	// [Context] values can then be referenced by key in the expression.
	// The subject is keyed by the value of global variable [SubjectKey],
	// which defaults to "@", but can be redefined at runtime.
	//
	// Nested objects can be referenced using dot notation.
	plan map[string]string

	// Scheme associates a [plan] with a contextual subject.
	//
	// A new scheme is evaluated for each [parse.Subject] in a given namespace.
	scheme struct {
		subject string
		plan    plan
	}

	// Domain associates a scheme with the namespace lineage that produced it.
	//
	// Every domain is a unique sequence of namespaces along with a subject
	// with which is is evaluated.
	//
	// The lineage is a list of namespaces that are used to produce the domain.
	// The scheme is a map of environment variables to evaluable expressions.
	//
	// A domain is created for each [parse.Namespace] in a given namespace.
	domain struct {
		path   lineage
		scheme scheme
	}
)
