package bus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	mb := New()

	assert.NotNil(t, mb)
	assert.NotNil(t, mb.inbound)
	assert.NotNil(t, mb.outbound)
	assert.NotNil(t, mb.handlers)
	assert.NotNil(t, mb.subscribers)
}

func TestPublishInbound(t *testing.T) {
	mb := New()

	msg := InboundMessage{
		Channel:  "telegram",
		SenderID: "user123",
		Content:  "Hello",
	}

	// Publish should not block
	mb.PublishInbound(msg)

	// Consume the message
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.ConsumeInbound(ctx)
	require.True(t, ok)
	assert.Equal(t, "telegram", got.Channel)
	assert.Equal(t, "user123", got.SenderID)
	assert.Equal(t, "Hello", got.Content)
}

func TestTryPublishInbound(t *testing.T) {
	mb := New()

	msg := InboundMessage{Channel: "telegram", Content: "Test"}

	// Should succeed
	ok := mb.TryPublishInbound(msg)
	assert.True(t, ok)
}

func TestConsumeInboundCanceled(t *testing.T) {
	mb := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, ok := mb.ConsumeInbound(ctx)
	assert.False(t, ok)
}

func TestPublishOutbound(t *testing.T) {
	mb := New()

	msg := OutboundMessage{
		Channel: "telegram",
		ChatID:  "chat123",
		Content: "Hello back",
	}

	mb.PublishOutbound(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.SubscribeOutbound(ctx)
	require.True(t, ok)
	assert.Equal(t, "telegram", got.Channel)
	assert.Equal(t, "chat123", got.ChatID)
	assert.Equal(t, "Hello back", got.Content)
}

func TestTryPublishOutbound(t *testing.T) {
	mb := New()

	msg := OutboundMessage{Channel: "telegram", Content: "Test"}

	ok := mb.TryPublishOutbound(msg)
	assert.True(t, ok)
}

func TestSubscribeOutboundCanceled(t *testing.T) {
	mb := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, ok := mb.SubscribeOutbound(ctx)
	assert.False(t, ok)
}

func TestRegisterHandler(t *testing.T) {
	mb := New()

	handler := func(InboundMessage) error { return nil }
	mb.RegisterHandler("telegram", handler)

	got, ok := mb.GetHandler("telegram")
	require.True(t, ok)
	assert.NotNil(t, got)
}

func TestGetHandlerNotFound(t *testing.T) {
	mb := New()

	_, ok := mb.GetHandler("nonexistent")
	assert.False(t, ok)
}

func TestSubscribe(t *testing.T) {
	mb := New()

	var received Event
	handler := func(event Event) {
		received = event
	}

	mb.Subscribe("sub1", handler)

	// Broadcast should deliver to subscriber
	mb.Broadcast(Event{Name: "test", Payload: "data"})
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, "test", received.Name)
	assert.Equal(t, "data", received.Payload)
}

func TestUnsubscribe(t *testing.T) {
	mb := New()

	var called bool
	handler := func(Event) { called = true }

	mb.Subscribe("sub1", handler)
	mb.Unsubscribe("sub1")

	mb.Broadcast(Event{Name: "test"})
	time.Sleep(10 * time.Millisecond)

	assert.False(t, called)
}

func TestBroadcastMultipleSubscribers(t *testing.T) {
	mb := New()

	var wg sync.WaitGroup
	wg.Add(2)

	var mu sync.Mutex
	count := 0

	handler1 := func(Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	}
	handler2 := func(Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	}

	mb.Subscribe("sub1", handler1)
	mb.Subscribe("sub2", handler2)

	mb.Broadcast(Event{Name: "test"})
	wg.Wait()

	assert.Equal(t, 2, count)
}

func TestBroadcastNonBlocking(t *testing.T) {
	mb := New()

	// Handler that blocks should not block Broadcast
	handler := func(Event) {
		time.Sleep(100 * time.Millisecond)
	}

	mb.Subscribe("sub1", handler)

	// Should return immediately even if handler is slow
	mb.Broadcast(Event{Name: "test"})
}

func TestMessageBusClose(t *testing.T) {
	mb := New()
	mb.Close()

	// After close, ConsumeInbound returns zero value immediately
	// Channel receive from closed channel returns zero value with ok=false
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	msg, _ := mb.ConsumeInbound(ctx)
	// When channel is closed, receive returns zero value
	assert.Equal(t, InboundMessage{}, msg)
}
