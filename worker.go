// Package worker provides supervised, multi-threaded background workers for togo.
// Register a worker with a concurrency; the Manager runs a goroutine pool with
// panic recovery, exponential-backoff restarts, and graceful shutdown. Auto-starts
// when the plugin is installed (blank-imported); disable with WORKER_AUTOSTART=false.
//
// Install: `togo install togo-framework/worker`.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/togo-framework/togo"
)

// Func is a worker's body. Return nil to be restarted (e.g. one-shot poll), or
// block until ctx is cancelled for a long-running loop. A returned error or panic
// triggers a backoff restart.
type Func func(ctx context.Context) error

type spec struct {
	name        string
	concurrency int
	fn          Func
}

var (
	regMu    sync.Mutex
	registry []spec
)

// Register adds a worker with the given concurrency (call from init() or boot).
func Register(name string, concurrency int, fn Func) {
	if concurrency < 1 {
		concurrency = 1
	}
	regMu.Lock()
	registry = append(registry, spec{name, concurrency, fn})
	regMu.Unlock()
}

func init() {
	togo.RegisterProviderFunc("worker", togo.PriorityLate+30, func(k *togo.Kernel) error {
		m := NewManager(k.Log)
		k.Set("workers", m)
		if os.Getenv("WORKER_AUTOSTART") != "false" {
			m.Start()
		}
		return nil
	})
}

// Manager runs and supervises registered workers.
type Manager struct {
	log    *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

// NewManager creates a manager (Start runs the registered workers).
func NewManager(log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{log: log, ctx: ctx, cancel: cancel}
}

// Start launches every registered worker's pool and a signal watcher for graceful
// shutdown (SIGINT/SIGTERM). Safe to call once.
func (m *Manager) Start() {
	m.once.Do(func() {
		regMu.Lock()
		specs := append([]spec(nil), registry...)
		regMu.Unlock()
		total := 0
		for _, s := range specs {
			for i := 0; i < s.concurrency; i++ {
				m.wg.Add(1)
				go m.run(s, i)
				total++
			}
		}
		if total > 0 {
			m.log.Info("workers started", "workers", len(specs), "goroutines", total)
		}
		go m.watchSignals()
	})
}

// Stop cancels all workers and waits for them to drain.
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}

func (m *Manager) watchSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		m.log.Info("workers shutting down")
		m.cancel()
	case <-m.ctx.Done():
	}
}

func (m *Manager) run(s spec, idx int) {
	defer m.wg.Done()
	backoff := time.Second
	for {
		if m.ctx.Err() != nil {
			return
		}
		err := safeCall(m.ctx, s.fn)
		switch {
		case m.ctx.Err() != nil:
			return
		case err != nil:
			m.log.Error("worker crashed; restarting", "worker", s.name, "instance", idx, "err", err, "backoff", backoff)
			if !sleep(m.ctx, backoff) {
				return
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
		default:
			// Clean return (one-shot): reset backoff, brief pause to avoid a busy loop.
			backoff = time.Second
			if !sleep(m.ctx, 100*time.Millisecond) {
				return
			}
		}
	}
}

// safeCall runs fn, converting a panic into an error.
func safeCall(ctx context.Context, fn Func) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn(ctx)
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// FromKernel returns the worker manager from the kernel container.
func FromKernel(k *togo.Kernel) (*Manager, bool) {
	v, ok := k.Get("workers")
	if !ok {
		return nil, false
	}
	m, ok := v.(*Manager)
	return m, ok
}
