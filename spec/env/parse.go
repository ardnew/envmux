package env

import (
	"go/constant"
	"go/parser"
	"math/big"
	"reflect"

	"github.com/expr-lang/expr/ast"

	goast "go/ast"

	"github.com/ardnew/envmux/spec/env/vars"
)

type parameterType struct {
	vars.Env[any]
}

// Visit implements the ast.Visitor interface and replaces all identifier nodes
// matching the implicit parameter key with its constant value from p.Env.
// The nodes are replaced using expr-lang function [ast.Patch].
//
// The type of the constant value is resolved using the same type inferencing
// rules as primitive Go literals. Supported types:
//
//   - Nil
//   - Bool
//   - Int Int8 Int16 Int32 Int64
//   - Uint Uint8 Uint16 Uint32 Uint64
//   - Float Float64
//   - String
//   - Any
func (p parameterType) Visit(n *ast.Node) {
	var (
		val any
		id  *ast.IdentifierNode
		v   string
		ok  bool
	)

	if val, ok = p.Env[vars.ParameterKey]; !ok {
		return // no parameter key defined
	}

	if id, ok = (*n).(*ast.IdentifierNode); !ok ||
		id.Value != vars.ParameterKey {
		return
	}

	if v, ok = val.(string); !ok {
		return
	}

	if expr, err := parser.ParseExpr(v); err == nil {
		if lit, ok := expr.(*goast.BasicLit); ok {
			val = constant.Val(constant.MakeFromLiteral(v, lit.Kind, 0))

			switch tv := val.(type) {
			case *big.Int:
				val = tv.Int64()

			case *big.Float:
				val, _ = tv.Float64()

			case *big.Rat:
				val, _ = tv.Float64()
			}

			val = reflect.ValueOf(val).Interface()

			ast.Patch(n, &ast.ConstantNode{Value: val})
		}
	}
}
