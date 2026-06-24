<!-- togo-brand -->
<p align="center">
  <img src=".github/assets/togo-mark.svg" width="96" alt="togo" />
</p>
<h1 align="center">worker</h1>
<p align="center"><sub>part of the <a href="https://github.com/togo-framework">togo-framework</a> — the full-stack Go + React framework</sub></p>

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


---

## 💎 Premium sponsors

togo is proudly sponsored by **ID8 Media** and **One Studio**.

<p align="center">
  <a href="https://id8media.com"><img src=".github/assets/id8media.svg" height="44" alt="ID8 Media" /></a>
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <a href="https://one-studio.co"><img src=".github/assets/one-studio.jpeg" height="44" alt="One Studio" /></a>
</p>
