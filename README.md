# actionguard-go

[![CI](https://github.com/thaicn1712/actionguard-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/thaicn1712/actionguard-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/thaicn1712/actionguard-go.svg)](https://pkg.go.dev/github.com/thaicn1712/actionguard-go)
[![license](https://img.shields.io/github/license/thaicn1712/actionguard-go.svg?style=flat)](LICENSE)

Policy-as-code (OPA-style, deny-overrides, fail-closed) for AI agent tool calls, in Go. The Go port of [`actionguard`](https://crates.io/crates/actionguard) (Rust).

## Install

```bash
go get github.com/thaicn1712/actionguard-go
```

## Usage

```go
policies := actionguard.NewPolicySet().
    With(actionguard.NewAllowList("read_file", "list_dir")).
    With(actionguard.NewDenyList("shell_exec"))

decision := policies.Check(actionguard.NewToolCall("shell_exec", nil))
if !decision.Allowed {
    // decision.Reason explains why
}
```

Two rules, always:

- **Deny-overrides** — if any policy votes `VoteDeny`, the call is denied, no matter what else voted `VoteAllow`.
- **Fail-closed** — if no policy affirmatively votes `VoteAllow`, the call is denied by default. A tool call is never allowed just because nothing objected.

## Built-in policies

`AllowList`, `DenyList`, `ArgMatchesRegex` (deny unless a named string argument matches a pattern — e.g. keep a `write_file` tool inside `/tmp/`).

## Async policies

For checks that need network access — an external authorization service, an LLM-as-judge call — implement `AsyncPolicy` instead of `Policy`, and use `AsyncPolicySet`:

```go
type ExternalApproval struct{ client *approval.Client }

func (e ExternalApproval) Vote(ctx context.Context, call actionguard.ToolCall) (actionguard.Vote, string, error) {
    approved, err := e.client.Check(ctx, call.Name, call.Arguments)
    if err != nil {
        return actionguard.VoteAbstain, "", err
    }
    if approved {
        return actionguard.VoteAllow, "", nil
    }
    return actionguard.VoteDeny, "rejected by external approval service", nil
}
```

## Examples

```bash
go run ./examples/deny_overrides_allow
```

## Benchmarks

`go test -bench=.`:

| Scenario | Time |
|---|---|
| `PolicySet.Check`, 3 policies (2 allow-list entries + 2 deny-list entries) | ~15 ns |

## License

MIT
