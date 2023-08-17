package stream

import (
	"context"
	"time"
)

// ContextSend sends the result to the channel, or aborts if the context is
// done. Returs true if the result was sent.
func ContextSend[T any, C chan T](ctx context.Context, ch C, item T) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- item:
		return true
	}
}

// ContextSendWithTimeout sends the item to the channel, aborts if the
// context is done, or if the given timeout expires. True is returned if the
// item has been sent to the channel
func ContextSendWithTimeout[T any, C chan T](
	ctx context.Context, ch C, item T, timeout time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(timeout):
		return false
	case ch <- item:
		return true
	}
}
