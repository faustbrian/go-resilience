# Migration

The published v2.0.0 contract uses `github.com/faustbrian/go-resilience/v2`.
Migrate imports to `/v2`,
set `BudgetConfig.MaxScopes` when the default is not appropriate, and verify
that wrapped context and budget errors use direct-only v2 classification.
Configured metadata and policy identities must contain printable text;
diagnostic error and event constructors sanitize control and malformed bytes.

V1 and v2 scope and attempt context keys are distinct. V1-bound retry
and hedge executors cannot consume a v2 scope by changing only this import.
Published Retry v2.1.0 and Hedge v1.1.0 explicitly consume v2 scopes while
retaining their v1 routes. Upgrade those focused executors before migrating
an active shared-budget composition; retain historical v1 tests as historical
variants. Contexts carrying both scope versions are rejected before dispatch.

## v1.1.0

Upgrade to Go 1.27.0 or newer before adopting v1.1.0. The public API,
error classification, and runtime contracts remain unchanged.

The benchmark comparator now requires Failsafe-Go v0.9.7. If an
application also imports Failsafe directly, Go module selection can raise
its version. Review the upstream budget-method replacements and stricter
adaptive-quantile validation in the
[v0.9.7 source](https://github.com/failsafe-go/failsafe-go/tree/v0.9.7).
Those upstream paths are not used by the resilience benchmark.

## Composition

Adopt composition without changing focused policy behavior first:

1. Create bounded `Metadata` at the transport or use-case boundary.
2. Adapt existing focused policies to `Policy[T]` and preserve their public
   compatibility APIs.
3. Put logical retry and hedge policies before attempt-scoped breaker,
   bulkhead, rate, and adaptive policies.
4. Create one `WorkBudgetScope` per logical call and attach its returned
   context.
5. Remove focused retry and hedge budget state only after parity tests prove
   both draw from the shared scope.
6. Enable bounded observation and compare local rejection, downstream failure,
   and amplification metrics before rollout.

Do not replace application deadlines with a generic timeout wrapper. Do not
make a non-idempotent operation replayable merely because a retry policy can
invoke it repeatedly.
