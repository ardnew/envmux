package parse

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/ardnew/envmux/pkg"
)

// Token is the unsigned integer token type used by the generated PEG parser.
// It encodes rule indices and positions within the parser buffer.
type Token uint32

const (
	// TokenSize is the size in bytes of a single [Token] value.
	TokenSize = int(unsafe.Sizeof(Token(0)))
	// DefaultBufSize is the default parser buffer size used when reading
	// manifests via [AST.ReadFrom].
	DefaultBufSize = 1 << 15 // 32 KiB

	// FS is the field separator character used when rendering manifests.
	FS = `,`
	// RS is the statement/record separator character used when rendering
	// manifests.
	RS = `;`
	// NL is the newline escape sequence used when rendering manifests.
	NL = `\n`

	co, cc = `<`, `>`
	so, sc = `{`, `}`
	po, pc = `(`, `)`
	ao, ac = `[`, `]`
)

// AST is the root node of the parsed manifest file.
type AST struct {
	parser[Token]

	bufSize int
	pretty  bool
}

// New constructs a new, empty [AST] with reasonable defaults applied.
// Use [WithBufSize] and [WithPretty] to override defaults as needed.
//
//nolint:exhaustruct
func New() *AST {
	return &AST{
		parser:  parser[Token]{},
		bufSize: DefaultBufSize,
		pretty:  true,
	}
}

// WithBufSize sets the internal parser buffer size used when reading manifests
// with [AST.ReadFrom]. Values smaller than [TokenSize] are clamped, and values
// are rounded up to the nearest multiple of [TokenSize].
func WithBufSize(bufSize int) pkg.Option[AST] {
	if bufSize < TokenSize {
		bufSize = DefaultBufSize // Sanity barrier.
	} else if bufSize%TokenSize != 0 {
		bufSize += TokenSize - (bufSize % TokenSize)
	}

	return func(a AST) AST {
		a.bufSize = bufSize

		return a
	}
}

// WithPretty toggles pretty formatting for [AST] when rendered via
// [fmt.Formatter] or [fmt.Stringer].
func WithPretty(pretty bool) pkg.Option[AST] {
	return func(a AST) AST {
		a.pretty = pretty

		return a
	}
}

// Format implements [fmt.Formatter] to render the AST in compact or pretty
// form, depending on the receiver's Pretty setting and the format verb.
func (a *AST) Format(f fmt.State, c rune) {
	if a == nil {
		fmt.Fprint(f, "<nil>")

		return
	}

	var sep, pad string
	if a.pretty {
		sep = "\n"

		if mult, ok := f.Width(); ok && f.Flag('-') {
			pad = strings.Repeat(" ", mult)
		}

		fmt.Fprint(f, "AST(", sep)
	}

	if strings.ContainsRune("sv", c) {
		for _, ns := range a.Namespaces {
			fmt.Fprint(f, pad, ns, sep)
		}
	}

	if a.pretty {
		fmt.Fprint(f, ")", sep)
	}
}

func (a *AST) String() string {
	if a == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%s", a)
}

const useBufferedReader = false

// ReadFrom parses a manifest read from the given [io.Reader] and populates the
// receiver [AST].
//
// ReadFrom implements [io.ReaderFrom] for maximum control and compatibility.
//
// For example, when combined with an appropriate [io.WriterTo] that produces
// a manifest, [io.Copy] can construct an [AST] from unbuffered
// and/or unseekable byte streams including sockets, pipes, mmaps, etc.
func (a *AST) ReadFrom(r io.Reader) (int64, error) {
	b := new(bytes.Buffer)

	if useBufferedReader {
		r = bufio.NewReader(r)
	}

	n, err := b.ReadFrom(r)
	if err != nil {
		return n, err
	}

	a.Buffer = b.String()

	options := []func(*parser[Token]) error{
		Pretty[Token](a.pretty),
		Size[Token](min(int(n), a.bufSize)),
	}

	err = a.Init(options...)
	if err != nil {
		return n, err
	}

	wrapParseError := func(err error) error {
		var errParse *parseError[Token]

		if errors.As(err, &errParse) {
			return pkg.MakeParseError(
				errParse.p.Buffer,
				int(errParse.maxToken.begin+1),
			)
		}

		return err
	}

	err = wrapParseError(a.Parse())
	defer a.Execute()

	return n, err
}

// ReadFile parses a manifest read from the given file path and populates the
// receiver [AST].
func (a *AST) ReadFile(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return a.ReadFrom(f)
}
