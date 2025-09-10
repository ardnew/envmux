package manifest

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/carlmjohnson/flowmatic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/file"

	"github.com/ardnew/envmux/manifest/builtin"
	"github.com/ardnew/envmux/manifest/config"
	"github.com/ardnew/envmux/manifest/parse"
	"github.com/ardnew/envmux/pkg"
	"github.com/ardnew/envmux/pkg/fn"
)

// Indirect references to functions to allow for testing.
//
//nolint:gochecknoglobals
var (
	compileExpr          = expr.Compile
	manifestFromStringFn = manifestFromString
)

// Model evaluates namespaced environment variables with expression support.
// Parses and evaluates environment manifests from a custom file format.
type Model struct {
	*parse.AST `json:"namespaces"`

	// Maximum number of jobs that may be run simultaneously.
	MaxParallelJobs int `json:"jobs,omitempty"`

	// Whether the model treats undefined namespaces as errors.
	StrictDefinitions bool `json:"requires,omitempty"`

	// ManifestReader is the reader used to read all manifests combined.
	ManifestReader io.Reader `json:"-"`
}

type parameterEnv struct {
	eval builtin.Env[any]
	pars []any
}

func export(sub ...parameterEnv) pkg.Option[parameterEnv] {
	return func(env parameterEnv) parameterEnv {
		return pkg.Wrap(
			env,
			func(t parameterEnv) parameterEnv {
				for _, o := range sub {
					maps.Copy(t.eval, o.eval)
					t.pars = append(t.pars, o.pars...)
				}

				return t
			},
		)
	}
}

// String returns the JSON representation of the model or an error message if
// marshaling fails.
func (m Model) String() string {
	e, err := json.Marshal(m)
	if err == nil {
		return pkg.JoinErrors(pkg.ErrInvalidJSON, err).Error()
	}

	return string(e)
}

// IsZero reports whether the model has no AST configured.
func (m Model) IsZero() bool { return m.AST == nil }

// Parse reads a manifest from [Model.ManifestReader] and initializes
// the returned [Model.AST] used to evaluate constructed environments.
// Parse reads a manifest from [Model.ManifestReader] and initializes the
// [Model.AST] used for evaluation.
func (m Model) Parse() (Model, error) {
	ast := parse.New()
	if _, err := ast.ReadFrom(m.ManifestReader); err != nil { //nolint:noinlineerr
		return Model{}, err
	}

	return pkg.Wrap(m, WithAST(ast)), nil
}

// Eval evaluates the requested namespaces and returns a fully constructed
// environment mapping. When [Model.StrictDefinitions] is true, unknown
// namespaces return an error.
func (m Model) Eval(
	ctx context.Context, namespaces ...string,
) (builtin.Env[any], error) {
	s := slices.Collect(fn.Filter(slices.Values(namespaces),
		func(ns string) bool { return ns != "" },
	))

	list := make([]parse.Composite, len(s))
	for i, id := range s {
		list[i] = parse.Composite{Ident: id, Parameters: nil}
	}

	env, err := m.eval(ctx, list...)

	return env.eval, err
}

func (m Model) eval(
	ctx context.Context, composites ...parse.Composite,
) (parameterEnv, error) {
	if len(composites) == 0 {
		return parameterEnv{}, nil //nolint:exhaustruct
	}

	maxJobs := min(len(composites), runtime.NumCPU())
	if m.MaxParallelJobs <= 0 {
		m.MaxParallelJobs = maxJobs
	}

	maxJobs = min(m.MaxParallelJobs, maxJobs)

	env := parameterEnv{
		eval: builtin.Env[any]{},
		pars: []any{},
	}

	// If we have only one job, run the evaluations serially in this goroutine
	// and don't fan out additional goroutines.
	if m.MaxParallelJobs == 1 {
		for _, co := range composites {
			e, err := m.evalComposition(ctx, co)
			if err != nil {
				return parameterEnv{}, err
			}

			env = pkg.Wrap(env, export(e))
		}
	} else {
		// Evaluate all namespaces in parallel,
		// returning a slice of fully-evaluated environments.
		for group := range slices.Chunk(composites, maxJobs) {
			envs, err := flowmatic.Map(ctx, len(group), group, m.evalComposition)
			if err != nil {
				return parameterEnv{}, err
			}

			env = pkg.Wrap(env, export(envs...))
		}
	}

	return env, nil
}

// FindDuplicateNamespaces is a debug option that panics on the
// detection of duplicate namespace definitions in the manifest.
//
// These most likely occur when a namespace is composed of itself
// indirectly or when the same manifest is included multiple times.
//
// These should never occur in practice because duplicate namespace
// instances in the composition graph are pruned prior to evaluation.
//
//nolint:gochecknoglobals
var findDuplicateNamespaces = false

func (m Model) evalComposition(
	ctx context.Context,
	composite parse.Composite,
) (parameterEnv, error) {
	// locate namespace by identifier
	idx := slices.IndexFunc(m.Namespaces, func(ns parse.Namespace) bool {
		return ns.Ident == composite.Ident
	})

	// Verify that the namespace exists in the model.
	if idx < 0 {
		if m.StrictDefinitions {
			return parameterEnv{}, pkg.ErrUndefinedNamespace.WithDetail(
				composite.Ident,
			)
		}

		return parameterEnv{}, nil //nolint:exhaustruct
	}

	def := m.Namespaces[idx]

	// optional duplicate detection (debug)
	if findDuplicateNamespaces {
		checkDuplicateDefinitions(m.Namespaces, idx, def)
	}

	env := parameterEnv{
		eval: builtin.Env[any]{},
		pars: []any{},
	}

	// recursively evaluate composed namespaces
	if len(def.Composites) > 0 {
		var err error

		env, err = m.eval(ctx, def.Composites...)
		if err != nil {
			return parameterEnv{}, err
		}
	}

	// collect parameters (definition, composed, inline)
	evalParams := slices.Concat(
		slices.Collect(def.Arguments()),
		env.pars,
		slices.Collect(composite.Arguments()),
	)
	if len(evalParams) == 0 {
		evalParams = append(evalParams, builtin.NoParameter)
	}

	// evaluate for each parameter set
	for _, par := range evalParams {
		e := pkg.Make(
			builtin.WithContext(ctx),
			// Calling [vars.WithParameter] when par == [vars.NoParameter] causes
			// [vars.ParameterKey] to be removed from the environment.
			builtin.WithParameter(par),
			builtin.WithExports(env.eval),
		)

		opt := []expr.Option{
			expr.Env(e.AsMap()),
			expr.Optimize(true),
			expr.WithContext(builtin.ContextKey),
			expr.AllowUndefinedVariables(),
			parameterType{e}.Patch(),
		}
		opt = append(opt, builtin.CacheCoerceConst()...)

		err := evalNamespaceStatements(def, e, opt, &env)
		if err != nil {
			return parameterEnv{}, err
		}
	}

	return env, nil
}

// checkDuplicateDefinitions detects duplicate namespace definitions and panics
// when duplicates are found (used only for debug mode).
func checkDuplicateDefinitions(
	namespaces []parse.Namespace,
	idx int,
	def parse.Namespace,
) {
	// Now that it has been resolved, we can search for duplicate definitions
	// (and not just duplicate identifiers).
	matchDups := func(s string) func(ns parse.Namespace) bool {
		return func(ns parse.Namespace) bool {
			return ns.String() == s
		}
	}(def.String())

	pos, dup := idx, make([]int, 0, len(namespaces)-idx-1)

	for 0 <= pos && pos < len(namespaces)-1 {
		off := pos + 1
		if pos = slices.IndexFunc(namespaces[off:], matchDups); pos >= 0 {
			dup = append(dup, off+pos)
			// advance absolute position to continue scanning after the found match
			pos = off + pos

			continue
		}

		break
	}

	if numDuplicates := len(dup); numDuplicates > 0 {
		panic(fmt.Sprintf(
			"found %d duplicate definitions for namespace:\n\t%s\n",
			numDuplicates,
			def.String(),
		))
	}
}

// evalNamespaceStatements compiles and runs each statement in the namespace
// for the provided environment and merges the results into envPtr.
func evalNamespaceStatements(
	def parse.Namespace,
	e builtin.Env[any],
	opt []expr.Option,
	envPtr *parameterEnv,
) error {
	wrapEvalError := func(sta parse.Statement, err error) error {
		var errFile *file.Error
		if errors.As(err, &errFile) {
			return pkg.MakeEvalError(
				def.Ident,
				sta.Ident,
				sta.Expression.Src,
				errFile.From+1,
			)
		}

		return err
	}

	for _, sta := range def.Statements {
		program, err := compileExpr(sta.Expression.Src, opt...)
		if err != nil {
			return wrapEvalError(sta, err)
		}

		res, err := expr.Run(program, e.AsMap())
		if err != nil {
			return err
		}

		e[sta.Ident] = unquote(res)
		maps.Copy(envPtr.eval, collect(e))
	}

	return nil
}

// Make constructs a [Model] by reading one or more manifest sources and
// applying the provided options. It accepts both file paths and inline
// definitions.
func Make(
	_ context.Context,
	manifests, defines []string,
	opts ...pkg.Option[Model],
) (Model, error) {
	manifest := make([]io.Reader, 0, len(manifests))

	nonEmpty := func(t string) (string, bool) {
		if t = strings.TrimSpace(t); t == "" {
			return "", false
		}

		return t, true
	}

	for path := range fn.Map(slices.Values(manifests), nonEmpty) {
		r, err := manifestFromPath(path)
		if err != nil {
			return Model{}, err
		}

		manifest = append(manifest, r)
	}

	for def := range fn.Map(slices.Values(defines), nonEmpty) {
		r, err := manifestFromStringFn(def)
		if err != nil {
			return Model{}, err
		}

		manifest = append(manifest, r)
	}

	return pkg.Make(
		append(opts, WithManifestReader(io.MultiReader(manifest...)))...), nil
}

// WithAST is a functional [pkg.Option] that installs the manifest
// parsed from a [parse.AST].
//
// It is a required option that must be applied prior to evaluating
// environment variables with [Model.Eval].
func WithAST(ast *parse.AST) pkg.Option[Model] {
	return func(m Model) Model {
		m.AST = ast

		return m
	}
}

// WithParallelEvalLimit is a functional [pkg.Option] that sets the maximum
// number of parallel jobs to run when evaluating the environment.
//
// The default number of jobs is equal to the number of CPU cores available.
// A value of 0 means to use the default number of jobs.
func WithParallelEvalLimit(n int) pkg.Option[Model] {
	return func(m Model) Model {
		m.MaxParallelJobs = n

		return m
	}
}

// WithStrictDefinitions is a functional [pkg.Option] that sets whether the
// model treats undefined namespaces as errors.
func WithStrictDefinitions(b bool) pkg.Option[Model] {
	return func(m Model) Model {
		m.StrictDefinitions = b

		return m
	}
}

// WithManifestReader is a functional [pkg.Option] that sets the reader used to
// read all manifests combined.
func WithManifestReader(r io.Reader) pkg.Option[Model] {
	return func(m Model) Model {
		m.ManifestReader = r

		return m
	}
}

// readerFromFile returns a buffered reader from the given file name.
func readerFromFile(filename string) (io.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return bufio.NewReader(f), nil
}

func manifestFromPath(path string) (io.Reader, error) {
	// Handle special cases for path to manifest file:
	//
	//   1. If [run.StdinSpecPath] given as flag argument, use stdin
	//   2. If flag argument is a relative path, use the first existing:
	//     A. relative to CWD
	//     B. relative to the manifest directory
	if path == config.StdinManifestPath {
		// Read from stdin
		return os.Stdin, nil
	}

	var (
		r   io.Reader
		err error
	)

	// During the first iteration, control will either:
	//
	//   1. successfully construct a reader (r != nil, break loop), or
	//   2. fail to construct a reader (r == nil) with error (err != nil), and:
	//     A. the path is absolute, error persists (err != nil, break loop), or
	//     B. the path is relative to CWD, try to set path as an absolute path
	//        relative to the spec directory, and:
	//       a. absolute path failed, overwrite error (err != nil, break loop), or
	//       b. absolute path succeeded, clear error (err == nil, continue to 1).
	//         - The second iteration will terminate at either 1 or 2A
	//           because condition 2B (the "recursive" step) will be false,
	//           since path is now guaranteed to be absolute.
	for r == nil && err == nil {
		r, err = readerFromFile(path)
		if err != nil && !filepath.IsAbs(path) {
			path, err = filepath.Abs(filepath.Join(config.Dir(pkg.Name), path))
		}
	}

	return r, err
}

func manifestFromString(def string) (io.Reader, error) {
	// Use a string reader to read a manifest from a string
	return io.NopCloser(strings.NewReader(def)), nil
}

func collect(e builtin.Env[any]) builtin.Env[any] {
	return maps.Collect(
		fn.FilterKeys(builtin.Cache().Complement(e),
			func(key string) bool {
				return key != builtin.ContextKey && key != builtin.ParameterKey
			},
		),
	)
}

// Unquote returns the unquoted value of a string or byte slice.
// If the value is not a string or byte slice, it is returned unchanged.
func unquote(value any) any {
	var s string

	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case interface{ String() string }:
		s = v.String()
	}

	us, err := strconv.Unquote(s)
	if err == nil {
		return us
	}

	return value
}
