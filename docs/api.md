# API reference

The supported v2 canonical generated reference is:

```sh
go doc -all github.com/faustbrian/go-resilience/v2
```

Primary entry points:

- `NewExecutor[T]` validates and freezes outer-to-inner policy composition when
  given no more than `MaxPolicies` policies;
- `Executor.Execute` runs one logical call synchronously;
- `Executor.WithClock`, `WithTimeline`, and `WithObserver` return immutable
  executor copies;
- `Policy`, `Stage`, and `Execution.WithAttempt` are the focused-policy seam;
- `NewMetadata` and `NewAttempt` validate bounded printable logical and physical
  identity;
- `Success`, `Failure`, `LocalRejection`, `Ignored`, and `PolicyFailure` build
  typed results;
- `NewBudget`, `WorkBudget`, `WorkBudgetScope`, and `Permit` own shared retry
  and hedge accounting;
- `WithBudgetScope` lets a custom budget implementation attach its scope;
- `BudgetScopeFromContext` exposes only an explicitly attached logical scope;
- `AdmitAttempt` coordinates unique retry and hedge lineage and returns its
  exactly-once permit; and
- `AttemptFromContext` exposes the physical attempt already owned by an outer
  executor so an inner executor cannot double-account it.

`Budget.Start` rejects an already scoped context with
`ErrBudgetAlreadyAttached`; nested execution must reuse the attached scope
rather than create a second accounting owner.

`BudgetConfig.MaxScopes` bounds simultaneously retained logical executions and
defaults to `DefaultMaxBudgetScopes` when zero. Configuration dimensions cannot
exceed the exported hard caps. New scope admission fails closed with
`ReasonScopeLimit` while the configured bound is occupied.

`PolicyDescriptor.Scope` is executable ordering metadata: logical policies
must wrap attempt policies. A repeatable policy identity may occur more than
once only when every occurrence explicitly declares the same scope and
repeatability.

V2 classifies context sentinels and budget rejection details only when they are
direct values. It preserves wrapped operation errors without invoking arbitrary
error traversal methods.
