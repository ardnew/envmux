package stream

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/ardnew/envmux/pkg"
)

// Stage is a function that returns the next value of type T
// from the provided [Group].
//
// The Stage function is used to define a processing stage in a pipeline.
// It is called repeatedly to produce a single value for the next stage by
// consuming one or more values from the previous stage.
type Stage[T any] func() (T, error)

// Pipe applies a processing [Stage] as an [pkg.Option] for a [Group].
func (stage Stage[T]) Pipe(ctx context.Context) pkg.Option[Group[T]] {
	var (
		group  *errgroup.Group
		cancel context.CancelCauseFunc
	)

	ctx, cancel = context.WithCancelCause(ctx)
	group, ctx = errgroup.WithContext(ctx)

	return func(Group[T]) Group[T] {
		// Pipe has the only task that can emit values via channel writes,
		// close channels, or capture context cancellations.
		//
		// Thus, this is the only scope in which the channel is writable/closable.
		out := make(chan T)

		// The signature of Make coerces out to become a read-only channel.
		//
		// [Group]s are the only objects that retain a reference to the channel
		// via field [Group.Channel].
		//
		// And because [Group.Channel] is read-only, it ensures the channel cannot
		// be written or closed anywhere except in the task closure below.
		s := Make(group, cancel, out)

		task := func() error {
			defer close(out)

			for {
				// Always request the next value. The receiver is responsible for
				// reading the channel as many times as required and reporting errors.
				o, err := stage()
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case out <- o:
				}
			}
		}

		s.Go(task)

		return s
	}
}
