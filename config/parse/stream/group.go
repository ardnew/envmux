// Package stream provides an implementation for streaming generic values
// using a pipeline of concurrent processing [Stage]s.
//
// The primary control surface is [Group]. Implemented as a [errgroup.Group],
// it contains a read-only channel [Group.Chan] for consuming the stream.
//
// [Tokens] returns a concrete [Group] implementation and initial processing
// [Stage] that streams [Token]s from a [lexer.PeekingLexer].
// [Tokens] forms the first stage in the parsing pipeline.
package stream

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/ardnew/envmux/pkg"
)

const redoLen = 16

// Group is a wait group used to coordinate a stream of values with type T.
//
// Group includes an [errgroup.Group] to manage the Group lifecycle.
// Users should use [Group.Go] to start tasks that write to [Group.Chan]
// and then wait for completion with [Group.Wait].
//
// [Group.Wait] returns an error indicating whether the stream successfully
// generated all values or failed due to error or cancellation.
//
// Tasks writing to the [Group.Chan] should always
//
//  1. perform all write operations in a goroutine started with [Group.Go],
//  2. emit errors by returning them from the writing task,
//  3. close the channel before returning from the writing task, and
//  4. verify that the context is not done before writing to the channel.
//     - e.g., a successful read from channel Done of [context.Context]
//
// Calling [Group.Cancel] with an error cause will
// cancel [Group.Group]'s context with that error.
type Group[T any] struct {
	*errgroup.Group
	context.CancelCauseFunc
	Channel <-chan T

	// redo is a slice of values that were read from the channel
	// but not accepted by the receiver.
	//
	// When value(s) are unaccepted, they are enqueued to redo.
	// Subsequent calls to an Accept method will attempt to dequeue
	// from redo before reading from the channel.
	redo []T
}

// Make constructs a new [Group] from the provided [errgroup.Group],
// [context.CancelCauseFunc], and read-only output channel of type T.
func Make[T any](
	group *errgroup.Group,
	cancel context.CancelCauseFunc,
	channel <-chan T,
) Group[T] {
	return Group[T]{
		Group:           group,
		CancelCauseFunc: cancel,
		Channel:         channel,
	}
}

func (g *Group[T]) undo(v T) {
	if g.redo == nil {
		g.redo = make([]T, 0, redoLen)
	}

	g.redo = append(g.redo, v)
}

//nolint:ireturn
func (g *Group[T]) next() (T, bool) {
	if len(g.redo) == 0 {
		v, ok := <-g.Channel

		return v, ok
	}

	v := g.redo[0]
	g.redo = g.redo[1:]

	return v, true
}

// Accept consumes the next value from the receiver's [Group.Chan]
// and returns it if it is accepted by the provided predicate.
//
// The returned error will be nil if the value was accepted.
// Otherwise, [pkg.ErrClosedStream] is returned on unsuccessful channel read,
// and [pkg.ErrUnacceptableStream] is returned if the value was not accepted.
//
//nolint:ireturn
func (g *Group[T]) Accept(predicate func(T) bool) (T, error) {
	v, ok := g.next()
	if !ok {
		return v, pkg.ErrClosedStream
	}

	if !predicate(v) {
		g.undo(v) // Put the value back to redo.

		return v, pkg.ErrUnacceptableStream
	}

	return v, nil
}

// AcceptAny consumes the next value from the receiver's [Group.Chan] and
// returns the first value accepted by at least one of the provided predicates.
//
// The returned error will be nil if the value was accepted by any predicate.
// Otherwise, [pkg.ErrClosedStream] is returned on unsuccessful channel read,
// and [pkg.ErrUnacceptableStream] is returned if the value was not accepted.
//
// The predicates are evaluated in the order they are provided.
// Evaluation stops with the first accepted value.
//
//nolint:ireturn
func (g *Group[T]) AcceptAny(predicates ...func(T) bool) (T, error) {
	v, ok := g.next()
	if !ok {
		return v, pkg.ErrClosedStream
	}

	for _, p := range predicates {
		if p(v) {
			return v, nil
		}
	}

	g.undo(v) // Put the value back to redo.

	return v, pkg.ErrUnacceptableStream
}

// AcceptEach returns values consumed from the receiver's [Group.Chan]
// that are accepted with respect to the order of provided predicates.
//
// The channel must produce at least as many values as there are predicates.
//
// This method is intended to be used to capture a fixed sequence of values.
//
// All values consumed from the channel are appended to the returned slice.
// If the channel produces a sequence containing an unaccepted value,
// a slice containing the unaccepted value in tail position is returned
// with error [pkg.ErrUnacceptableStream].
//
// If the channel produces fewer values than the number of given predicates
// (e.g., closed channel), a slice containing all consumed values is returned
// with error [pkg.ErrClosedStream].
//
// Given zero predicates, the empty, non-nil slice with nil error is returned.
func (g *Group[T]) AcceptEach(predicates ...func(T) bool) ([]T, error) {
	a := make([]T, 0, len(predicates))

	for _, p := range predicates {
		v, ok := g.next()
		if !ok {
			return a, pkg.ErrClosedStream
		}

		// Append all values consumed from the channel,
		// regardless of whether they are accepted.
		a = append(a, v)

		if !p(v) {
			g.undo(v) // Put the value back to redo.

			return a, pkg.ErrUnacceptableStream
		}
	}

	return a, nil
}
