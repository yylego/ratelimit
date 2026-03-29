[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/ratelimit/release.yml?branch=main&label=BUILD)](https://github.com/yylego/ratelimit/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/ratelimit)](https://pkg.go.dev/github.com/yylego/ratelimit)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/ratelimit/main.svg)](https://coveralls.io/github/yylego/ratelimit?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.22+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/ratelimit.svg)](https://github.com/yylego/ratelimit/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/ratelimit)](https://goreportcard.com/report/github.com/yylego/ratelimit)

# ratelimit

In-process rate limiter with per-key independent throttling.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Features

- **Sliding Window**: Counts requests within a configurable window, precise QPS enforcement
- **Per-Key Isolation**: Each key gets its own independent limiter, no cross-contamination
- **Auto Eviction**: Idle keys are cleaned up to prevent unbounded RAM usage
- **Batch Sweep**: Configurable max evictions per sweep via `SetSweepBatch`
- **Background Sweep**: Goroutine-based sweep via `StartSweepGoroutine` / `CloseSweepGoroutine`
- **Concurrence Safe**: Mutex on key map + heap, independent mutex on each limiter instance

## Installation

```bash
go get github.com/yylego/ratelimit
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/yylego/ratelimit"
)

func main() {
    // create a group: 1000 requests per second per key
    gp := ratelimit.NewGroup(1000, time.Second)

    // check if request is allowed
    if gp.Allow("hot_name") {
        fmt.Println("request accepted")
    } else {
        fmt.Println("rate limited")
    }

    // custom window: 100 requests per minute per key
    gp2 := ratelimit.NewGroup(100, time.Minute)
    gp2.Allow("some_name")
}
```

## Design

### Sliding Window

Each `Limiter` maintains a sorted slice of request timestamps (nanoseconds). On each `Allow()`:

1. Binary search to trim entries outside the configured window
2. Check if count < threshold
3. Append current timestamp if within limit

### Auto Eviction

Idle keys are evicted via min-heap. Two modes:

- **Inline mode** (default): On each `Allow()` call, the heap top is checked and expired keys are removed, up to `sweepBatch` at a time
- **Background mode**: Call `StartSweepGoroutine(tick)` to launch a dedicated goroutine that sweeps on a fixed tick. `Allow()` skips inline sweep in this mode. Call `CloseSweepGoroutine()` to close and switch back

Map and heap are 1:1, no duplicate nodes.

### Concurrence

- `Group` mutex is held to access the key map and heap, then released before `Limiter.Allow()`
- Each `Limiter` uses its own `sync.Mutex` — different keys are rate-checked in true concurrence

## API

| Type      | Method                                    | Description                                   |
| --------- | ----------------------------------------- | --------------------------------------------- |
| `Group`   | `NewGroup(threshold, window)`             | Create per-key limiter group                  |
| `Group`   | `Allow(key) bool`                         | Check if request is within limit              |
| `Group`   | `SetSweepBatch(n)`                        | Set max evictions per sweep                   |
| `Group`   | `StartSweepGoroutine(tick)`               | Start background sweep goroutine              |
| `Group`   | `CloseSweepGoroutine()`                   | Close background sweep, switch back to inline |
| `Limiter` | `NewLimiter(threshold)`                   | Create single-key limiter (1s window)         |
| `Limiter` | `NewLimiterWithWindow(threshold, window)` | Create single-key limiter (custom window)     |
| `Limiter` | `Allow() bool`                            | Check if request is within limit              |

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->
