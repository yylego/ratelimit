[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/ratelimit/release.yml?branch=main&label=BUILD)](https://github.com/yylego/ratelimit/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/ratelimit)](https://pkg.go.dev/github.com/yylego/ratelimit)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/ratelimit/main.svg)](https://coveralls.io/github/yylego/ratelimit?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.22+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/ratelimit.svg)](https://github.com/yylego/ratelimit/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/ratelimit)](https://goreportcard.com/report/github.com/yylego/ratelimit)

# ratelimit

轻量级内存限流器，支持按 Key 独立限流。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 功能特性

- **滑动窗口**：统计过去 1 秒内的请求数，精确 QPS 控制
- **按 Key 隔离**：每个 Key 独立限流，互不干扰
- **自动过期**：长时间无请求的 Key 自动清理，防止内存无限增长
- **并发安全**：Key 映射表使用读写锁，每个限流器独立互斥锁

## 安装

```bash
go get github.com/yylego/ratelimit
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/yylego/ratelimit"
)

func main() {
    // 创建限流组：每个 Key 每秒最多 1000 次请求
    gp := ratelimit.NewGroup(1000, time.Second)

    // 检查请求是否允许
    if gp.Allow("hot_name") {
        fmt.Println("请求通过")
    } else {
        fmt.Println("被限流")
    }

    // 自定义窗口：每个 Key 每分钟最多 100 次
    gp2 := ratelimit.NewGroup(100, time.Minute)
    gp2.Allow("some_name")
}
```

## 设计

### 滑动窗口

每个 `Limiter` 维护一个有序的请求时间戳切片（纳秒精度）。每次 `Allow()` 调用时：

1. 二分查找裁剪窗口外的记录
2. 检查当前数量是否 < 阈值
3. 在限额内则追加当前时间戳

### 自动淘汰

空闲 Key 通过最小堆懒淘汰 — 无需后台协程。每次 `Allow()` 调用时检查堆顶，移除过期 Key。Map 和堆 1:1 对应，无重复节点。

### 并发设计

- `Group` 锁只保护 Key 映射表和堆操作，释放后再调 `Limiter.Allow()`
- 每个 `Limiter` 使用独立的 `sync.Mutex` — 不同 Key 的限流判断真正并行

## 接口

| 类型 | 方法 | 说明 |
|------|------|------|
| `Group` | `NewGroup(threshold, window)` | 创建按 Key 限流组 |
| `Group` | `Allow(key) bool` | 检查请求是否在限额内 |
| `Limiter` | `NewLimiter(threshold)` | 创建单 Key 限流器（1 秒窗口） |
| `Limiter` | `NewLimiterWithWindow(threshold, window)` | 创建单 Key 限流器（自定义窗口） |
| `Limiter` | `Allow() bool` | 检查请求是否在限额内 |

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->
