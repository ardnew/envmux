package manifest

import (
	"go/constant"
	"go/parser"
	"math/big"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"

	goast "go/ast"

	"github.com/ardnew/envmux/manifest/builtin"
)

type parameterType struct {
	builtin.Env[any]
}

// Patch returns an [expr.Option] that applies a patch to the expression AST
// using the [ast.Visitor] implemented by [parameterType.Visit].
// It is a thin wrapper that delegates to expr.Patch,
// providing a name that is better associated with its intent (versus Visit).
func (p parameterType) Patch() expr.Option { return expr.Patch(p) }

// Visit implements the [ast.Visitor] interface and replaces all identifier
// nodes matching the implicit parameter key with its constant value from p.Env.
// The nodes are replaced using expr-lang function [ast.Patch].
//
// The type of the constant value is resolved using the same type inferencing
// rules as primitive Go literals.
func (p parameterType) Visit(n *ast.Node) {
	v, ok := p.identifier(n)
	if !ok {
		return
	}

	exprNode, err := parser.ParseExpr(v)
	if err != nil {
		return
	}

	lit, ok := exprNode.(*goast.BasicLit)
	if !ok {
		return
	}

	if val, ok := coerceType(v, lit); ok {
		ast.Patch(n, &ast.ConstantNode{Value: val})
	}
}

func (p parameterType) identifier(n *ast.Node) (string, bool) {
	val, ok := p.Env[builtin.ParameterKey]
	if !ok {
		return "", false // no parameter key defined
	}

	id, ok := (*n).(*ast.IdentifierNode)
	if !ok || id.Value != builtin.ParameterKey {
		return "", false
	}

	v, ok := val.(string)
	if !ok {
		return "", false
	}

	return v, true
}

func coerceType(src string, lit *goast.BasicLit) (any, bool) {
	val := constant.Val(constant.MakeFromLiteral(src, lit.Kind, 0))

	switch tv := val.(type) {
	case bool:
		return tv, true
	case string:
		return tv, true
	case int64:
		return tv, true
	case *big.Int:
		return tv.Int64(), true
	case *big.Float:
		f, _ := tv.Float64()

		return f, true
	case *big.Rat:
		f, _ := tv.Float64()

		return f, true
	default:
		return reflect.ValueOf(val).Interface(), true
	}
}
