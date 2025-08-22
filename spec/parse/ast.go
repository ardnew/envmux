// Package parse implements a parser for manifests containing namespace
// definitions.
package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/ardnew/envmux/pkg/fn"
)

type Token uint32

const (
	TokenSize      = int(unsafe.Sizeof(Token(0)))
	DefaultBufSize = 1 << 15 // 32 KiB

	FS = `,`  // FS matches a field separator.
	RS = `;`  // RS matches a record separator.
	NL = `\n` // NL matches a newline character.

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

//nolint:exhaustruct
func New() *AST {
	return &AST{
		parser:  parser[Token]{},
		bufSize: DefaultBufSize,
		pretty:  true,
	}
}

func WithBufSize(bufSize int) fn.Option[AST] {
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

func WithPretty(pretty bool) fn.Option[AST] {
	return func(a AST) AST {
		a.pretty = pretty

		return a
	}
}

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

	if err = a.Init(options...); err != nil {
		return n, err
	}

	err = a.Parse()
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
