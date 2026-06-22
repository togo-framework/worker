# worker

Supervised, multi-threaded background workers for [togo](https://github.com/togo-framework/togo).
Register a worker with a concurrency, and the manager runs a goroutine pool with
panic recovery, backoff restarts, and graceful shutdown.

```bash
togo install togo-framework/worker
```

```go
import "github.com/togo-framework/worker"

func init() {
  worker.Register("emails", 4, func(ctx context.Context) error {
    // pull a job and process it; return to be restarted, or block until ctx done
    return nil
  })
}
```
