package parse

// Expression contains the text of an Expression recognized by the
// [github.com/expr-lang/expr] grammar.
type Expression struct {
	Src string
}

func (e Expression) String() string { return e.Src }
