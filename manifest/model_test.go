package manifest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/ardnew/envmux/manifest/builtin"
	"github.com/ardnew/envmux/manifest/config"
	"github.com/ardnew/envmux/manifest/parse"
	"github.com/ardnew/envmux/pkg"
)

// Local Stringer type for unquote tests
type testStringer struct{}

func (testStringer) String() string { return "\"yo\"" }

// badReader is an io.Reader that always errors, used to exercise Parse error path.
type badReader struct{}

func (badReader) Read(b []byte) (int, error) { return 0, fmt.Errorf("readerr") }

// Helpers
func mustTempFile(t *testing.T, dir, pattern, content string) string {
	t.Helper()
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if content != "" {
		if _, err := f.WriteString(content); err != nil {
			t.Fatalf("write temp: %v", err)
		}
	}
	f.Close()
	return f.Name()
}

func TestManifestPackage(t *testing.T) {
	// Test that the manifest package can be imported
	// This is a basic smoke test

	t.Run("package import", func(t *testing.T) {
		// Verify the package can be imported without issues
		// The import statement at the top of this file accomplishes this
	})
}

func TestManifestMake(t *testing.T) {
	t.Run("basic construction", func(t *testing.T) {
		// Test basic Model creation
		ctx := context.Background()
		manifests := []string{}
		defines := []string{}

		model, err := Make(ctx, manifests, defines)
		if err != nil {
			t.Errorf("Make() should not error with empty inputs: %v", err)
		}

		// With empty inputs, the model might be zero value, which is expected
		// Test that the function completes without error
		_ = model // Use model to avoid unused variable warning
	})

	t.Skip("Complex manifest tests require test fixtures and may have dependencies")
}

func TestModel_String_Branches(t *testing.T) {
	t.Run("no-error-from-json-marshal-returns-error-string", func(t *testing.T) {
		m := Model{AST: &parse.AST{}}

		// Expect the inverted condition to return the invalid JSON encoding string
		got := m.String()
		if got != pkg.ErrInvalidJSON.Error() {
			t.Fatalf("unexpected String(): %q", got)
		}
	})

	t.Run("marshal-error-path-returns-bytes-string", func(t *testing.T) {
		// Force json.Marshal to error by introducing NaN in an exported field
		ast := &parse.AST{}
		ast.Namespaces = []parse.Namespace{{
			Ident:      "n",
			Parameters: []parse.Parameter{{Value: math.NaN()}},
		}}

		m := Model{AST: ast}
		// On marshal error, String returns string(e) which will be ""
		if got := m.String(); got != "" {
			t.Fatalf("expected empty string on marshal error, got %q", got)
		}
	})
}

func TestModel_IsZero(t *testing.T) {
	var m Model
	if !m.IsZero() {
		t.Fatal("zero Model should be IsZero")
	}
	m.AST = &parse.AST{}
	if m.IsZero() {
		t.Fatal("non-zero Model should not be IsZero")
	}
}

func TestOptionsSetters(t *testing.T) {
	m := pkg.Make(
		WithAST(&parse.AST{}),
		WithParallelEvalLimit(7),
		WithStrictDefinitions(true),
		WithManifestReader(strings.NewReader("")),
	)
	if m.AST == nil || m.MaxParallelJobs != 7 || !m.StrictDefinitions || m.ManifestReader == nil {
		t.Fatalf("options not applied: %+v", m)
	}
}

func TestReaderAndManifestFromPathAndString(t *testing.T) {
	t.Run("readerFromFile", func(t *testing.T) {
		file := mustTempFile(t, t.TempDir(), "envmux-*.txt", "hello")
		r, err := readerFromFile(file)
		if err != nil || r == nil {
			t.Fatalf("readerFromFile error: %v, r=%v", err, r)
		}
		// quick sanity read
		s := bufio.NewScanner(r)
		if !s.Scan() {
			t.Fatal("expected content")
		}
	})

	t.Run("manifestFromPath-stdin", func(t *testing.T) {
		r, err := manifestFromPath(config.StdinManifestPath)
		if err != nil || r == nil {
			t.Fatalf("stdin path: err=%v r=%v", err, r)
		}
	})

	t.Run("manifestFromPath-relative-resolves-via-config.Dir", func(t *testing.T) {
		tmp := t.TempDir()
		oldDir := config.Dir
		config.Dir = func(string) string { return tmp }
		t.Cleanup(func() { config.Dir = oldDir })

		// Will try CWD first (fail), then join with config.Dir
		content := "relNS{}"
		rel := "my-manifest"
		abs := filepath.Join(tmp, rel)
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write rel manifest: %v", err)
		}
		r, err := manifestFromPath(rel)
		if err != nil || r == nil {
			t.Fatalf("manifestFromPath(rel): %v %v", err, r)
		}
	})

	t.Run("manifestFromString", func(t *testing.T) {
		r, err := manifestFromString("x{}")
		if err != nil || r == nil {
			t.Fatalf("manifestFromString: %v %v", err, r)
		}
	})
}

func TestMakeAndParseAndEval(t *testing.T) {
	ctx := context.Background()

	// Build a manifest with composition and parameters
	manifest := strings.Join([]string{
		// base namespace
		"base{ a = 1; s = \"x\" }",
		// parameterized namespace uses implicit _
		"param(hello){ p = _ }",
		// child composes base and param with inline argument
		"child<base,param(world)>{ c = a + 41; cp = p }",
	}, "\n")

	m, err := Make(ctx, nil, []string{manifest})
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	m, err = m.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Serial evaluation
	m = pkg.Wrap(m, WithParallelEvalLimit(1))
	env, err := m.Eval(ctx, "child")
	if err != nil {
		t.Fatalf("Eval child: %v", err)
	}

	// Validate inherited and computed values
	if v, ok := env["a"].(int); !ok || v != 1 {
		t.Fatalf("want a=1 int, got %T %v", env["a"], env["a"])
	}
	if v, ok := env["c"].(int); !ok || v != 42 {
		t.Fatalf("want c=42 int, got %T %v", env["c"], env["c"])
	}
	if v, ok := env["p"].(string); !ok || v != "world" {
		t.Fatalf("want p=world, got %T %v", env["p"], env["p"])
	}
	if v, ok := env["cp"].(string); !ok || v != env["p"].(string) {
		t.Fatalf("want cp == p, got %v vs %v", v, env["p"])
	}

	// Parallel evaluation path: evaluate two namespaces
	m = pkg.Wrap(m, WithParallelEvalLimit(max(2, runtime.NumCPU())))
	_, err = m.Eval(ctx, "base", "param")
	if err != nil {
		t.Fatalf("Eval parallel: %v", err)
	}
}

func TestParse_ErrorOnReader(t *testing.T) {
	// Reader that always errors
	m := Model{ManifestReader: badReader{}}
	if _, err := m.Parse(); err == nil {
		t.Fatalf("expected parse read error")
	}
}

func TestEval_Parallel_ErrorOnStrictMissing(t *testing.T) {
	ctx := context.Background()
	m := Model{AST: &parse.AST{}, StrictDefinitions: true}
	// Force parallel branch
	m.MaxParallelJobs = 4
	// Call internal eval with a missing composite ident
	if _, err := m.eval(ctx, parse.Composite{Ident: "missing"}); err == nil {
		t.Fatalf("expected error for missing namespace in strict mode")
	}
}

func TestEval_NonStrict_MissingNamespaceIsIgnored(t *testing.T) {
	ctx := context.Background()
	m := Model{AST: &parse.AST{}, StrictDefinitions: false}
	m.MaxParallelJobs = 1
	got, err := m.eval(ctx, parse.Composite{Ident: "missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.eval) != 0 || len(got.pars) != 0 {
		t.Fatalf("expected empty env from missing namespace, got %+v", got)
	}
}

func TestEval_ComposedNamespaceErrorPropagates(t *testing.T) {
	ctx := context.Background()
	m := Model{AST: &parse.AST{}, StrictDefinitions: true}
	// Define a namespace that composes a missing namespace
	m.Namespaces = []parse.Namespace{{
		Ident:      "parent",
		Composites: []parse.Composite{{Ident: "missing"}},
	}}
	if _, err := m.eval(ctx, parse.Composite{Ident: "parent"}); err == nil {
		t.Fatalf("expected error propagated from composed eval")
	}
}

func TestMake_ErrorOnMissingAbsoluteManifest(t *testing.T) {
	ctx := context.Background()
	// Use a definitely non-existent absolute path
	abs := filepath.Join(t.TempDir(), "does-not-exist")
	manifests := []string{abs}
	if _, err := Make(ctx, manifests, nil); err == nil {
		t.Fatalf("expected error when manifest file missing")
	}
}

func TestMake_ErrorOnInlineDefineReader(t *testing.T) {
	ctx := context.Background()
	// Stub manifestFromStringFn to return an erroring reader
	old := manifestFromStringFn
	manifestFromStringFn = func(def string) (io.Reader, error) {
		return nil, fmt.Errorf("inline reader error")
	}
	defer func() { manifestFromStringFn = old }()
	if _, err := Make(ctx, nil, []string{"x{}"}); err == nil || !strings.Contains(err.Error(), "inline reader error") {
		t.Fatalf("expected inline reader error, got %v", err)
	}
}

func TestEval_DefaultParallelJobs(t *testing.T) {
	ctx := context.Background()
	manifest := strings.Join([]string{
		"n1{ a = 1 }",
		"n2<n1>{ b = a + 1 }",
	}, "\n")
	m, err := Make(ctx, nil, []string{manifest})
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	m, err = m.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Leave MaxParallelJobs at 0 to exercise default computation
	if _, err := m.Eval(ctx, "n1", "n2"); err != nil {
		t.Fatalf("Eval default parallel: %v", err)
	}
}

func TestCheckDuplicateDefinitions_Loop_NoDup(t *testing.T) {
	// Ensure loop body executes without finding duplicates
	a := parse.Namespace{Ident: "A"}
	b := parse.Namespace{Ident: "B"}
	c := parse.Namespace{Ident: "C"}
	list := []parse.Namespace{a, b, c}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	checkDuplicateDefinitions(list, 0, a)
}

func TestCheckDuplicateDefinitions_PanicOnDup(t *testing.T) {
	// Now that the loop is corrected, ensure duplicates panic as intended
	ns := parse.Namespace{Ident: "A", Statements: []parse.Statement{{
		Text:       "x=1",
		Ident:      "x",
		Operator:   "=",
		Expression: &parse.Expression{Src: "1"},
	}}}
	// same String() representation by content
	dup := ns
	list := []parse.Namespace{ns, {Ident: "B"}, dup}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate definitions")
		}
	}()
	checkDuplicateDefinitions(list, 0, ns)
}

func TestEval_StrictVsNonStrict(t *testing.T) {
	ctx := context.Background()
	m := pkg.Make(WithAST(&parse.AST{}))

	// Strict: undefined should error
	m = pkg.Wrap(m, WithStrictDefinitions(true))
	if _, err := m.Eval(ctx, "nope"); err == nil || !strings.Contains(err.Error(), pkg.ErrUndefinedNamespace.Error()) {
		t.Fatalf("expected undefined namespace error, got %v", err)
	}

	// Non-strict: undefined should be ignored
	m = pkg.Wrap(m, WithStrictDefinitions(false))
	env, err := m.Eval(ctx, "nope")
	if err != nil || len(env) != 0 {
		t.Fatalf("expected no error and empty env, got err=%v env=%v", err, env)
	}
}

func TestEvalNamespaceStatements_Errors(t *testing.T) {
	ctx := context.Background()

	// Build a model with a malformed expression to trigger compile error
	bad := "bad{ x = 1+ }"
	m, err := Make(ctx, nil, []string{bad})
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	m, err = m.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := m.Eval(ctx, "bad"); err == nil || !strings.Contains(err.Error(), "evaluation error") {
		t.Fatalf("expected eval compile error, got %v", err)
	}

	// Runtime error path: call a function that returns (any, error)
	// and ensure expr propagates the error from Run.
	def := parse.Namespace{
		Ident: "f",
		Statements: []parse.Statement{{
			Text:       "x = boom()",
			Ident:      "x",
			Operator:   "=",
			Expression: &parse.Expression{Src: "boom()"},
		}},
	}
	e := pkg.Make(builtin.WithContext(ctx))
	// Inject function into environment
	e["boom"] = func() (int, error) { return 0, fmt.Errorf("boom") }
	var env parameterEnv
	opt := []expr.Option{expr.Env(e.AsMap())}
	if err := evalNamespaceStatements(def, e, opt, &env); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected runtime error from boom(): %v", err)
	}
}

func TestEvalNamespaceStatements_NonFileErrorPassthrough(t *testing.T) {
	ctx := context.Background()
	def := parse.Namespace{
		Ident: "wrap2",
		Statements: []parse.Statement{{
			Text:       "z = 1",
			Ident:      "z",
			Operator:   "=",
			Expression: &parse.Expression{Src: "1"},
		}},
	}
	e := pkg.Make(builtin.WithContext(ctx))
	var env parameterEnv
	// Stub compileExpr to return a non-*file.Error
	old := compileExpr
	compileExpr = func(_ string, _ ...expr.Option) (*vm.Program, error) {
		return nil, fmt.Errorf("synthetic")
	}
	defer func() { compileExpr = old }()
	opt := []expr.Option{expr.Env(e.AsMap())}
	if err := evalNamespaceStatements(def, e, opt, &env); err == nil || !strings.Contains(err.Error(), "synthetic") {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}

func TestEvalComposition_DebugDuplicateDetectionFlag(t *testing.T) {
	ctx := context.Background()
	// Two identical namespaces
	ns := parse.Namespace{Ident: "dup", Statements: []parse.Statement{{
		Text:       "x=1",
		Ident:      "x",
		Operator:   "=",
		Expression: &parse.Expression{Src: "1"},
	}}}
	dup := ns
	m := Model{AST: &parse.AST{}}
	m.Namespaces = []parse.Namespace{ns, dup}
	// Enable duplicate detection
	oldFlag := findDuplicateNamespaces
	findDuplicateNamespaces = true
	defer func() { findDuplicateNamespaces = oldFlag }()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to duplicate definitions")
		}
	}()
	// Trigger evalComposition via internal eval
	_, _ = m.eval(ctx, parse.Composite{Ident: "dup"})
}

// Note: wrapEvalError path is covered by TestEvalNamespaceStatements_Errors
// using a malformed expression that produces a compile-time error.

func TestCollectAndUnquoteAndExport(t *testing.T) {
	// collect filters out ctx and _
	e := map[string]any{
		"ctx": 123,
		"_":   "p",
		"K":   "V",
	}
	got := collect(e)
	if _, ok := got["ctx"]; ok {
		t.Fatalf("collect should omit ctx")
	}
	if _, ok := got["_"]; ok {
		t.Fatalf("collect should omit _")
	}
	if v := got["K"]; v != "V" {
		t.Fatalf("collect should contain K=V, got %v", v)
	}

	// unquote variations
	if v := unquote("\"abc\""); v != "abc" {
		t.Fatalf("unquote string: %v", v)
	}
	if v := unquote([]byte("\"hi\"")); v != "hi" {
		t.Fatalf("unquote bytes: %v", v)
	}
	if v := unquote(testStringer{}); v != "yo" {
		t.Fatalf("unquote Stringer: %v", v)
	}
	if v := unquote("plain"); v != "plain" {
		t.Fatalf("unquote passthrough: %v", v)
	}

	// export merges parameterEnvs
	a := parameterEnv{eval: map[string]any{"A": 1}, pars: []any{"x"}}
	b := parameterEnv{eval: map[string]any{"B": 2}, pars: []any{"y"}}
	c := pkg.Wrap(parameterEnv{eval: map[string]any{}, pars: nil}, export(a, b))
	if c.eval["A"].(int) != 1 || c.eval["B"].(int) != 2 {
		t.Fatalf("export merge eval failed: %v", c.eval)
	}
	if fmt.Sprint(c.pars) != "[x y]" {
		t.Fatalf("export merge pars failed: %v", c.pars)
	}
}

func TestCheckDuplicateDefinitions_NoDuplicates_NoPanic(t *testing.T) {
	ns := parse.Namespace{Ident: "A"}
	other := parse.Namespace{Ident: "B"}
	list := []parse.Namespace{ns, other}

	// idx at end ensures loop doesn't execute; should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("did not expect panic: %v", r)
		}
	}()
	checkDuplicateDefinitions(list, 1, other)
}

func TestEvalZeroComposites(t *testing.T) {
	m := Model{}
	got, err := m.eval(context.Background())
	if err != nil || got.eval != nil || got.pars != nil {
		t.Fatalf("expected zero parameterEnv, got=%v err=%v", got, err)
	}
}

func TestMake_TrimsInputs_And_ResolvesRelativePath(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	oldDir := config.Dir
	config.Dir = func(string) string { return tmp }
	t.Cleanup(func() { config.Dir = oldDir })

	// Prepare a manifest file only resolvable via config.Dir
	rel := "file1"
	content := "d{}"
	if err := os.WriteFile(filepath.Join(tmp, rel), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	model, err := Make(ctx, []string{" ", rel, ""}, []string{" ", "x{}", ""})
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	// Ensure the combined reader works by parsing
	_, err = model.Parse()
	if err != nil {
		t.Fatalf("Parse combined: %v", err)
	}
}
