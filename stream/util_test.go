package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextSend(t *testing.T) {
	ch := make(chan bool)
	defer close(ch)

	var sent bool

	go func() {
		sent = ContextSend(context.Background(), ch, true)
		ch <- true
	}()

	assert.True(t, <-ch)
	<-ch
	assert.True(t, sent)
}

func TestContextSendWithTimeout(t *testing.T) {
	ch := make(chan bool)
	defer close(ch)

	sent := ContextSendWithTimeout(
		context.Background(), ch, true, 5*time.Millisecond)

	assert.False(t, sent)
}

func TestContextSendCancel(t *testing.T) {
	ch := make(chan bool)
	defer close(ch)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.False(t, ContextSend(ctx, ch, true))
}
