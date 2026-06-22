package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerRunsAndStops(t *testing.T) {
	var started int32
	Register("test", 3, func(ctx context.Context) error {
		atomic.AddInt32(&started, 1)
		<-ctx.Done()
		return ctx.Err()
	})
	m := NewManager(nil)
	m.Start()
	time.Sleep(80 * time.Millisecond)
	m.Stop() // cancels + waits for drain
	if got := atomic.LoadInt32(&started); got != 3 {
		t.Fatalf("expected 3 worker instances, got %d", got)
	}
}

func TestSafeCallRecoversPanic(t *testing.T) {
	err := safeCall(context.Background(), func(context.Context) error { panic("boom") })
	if err == nil {
		t.Fatal("expected panic to be converted to an error")
	}
}
