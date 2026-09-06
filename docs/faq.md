# FAQ

## Is this a resilience framework?

No. It is a small explicit composition contract used by focused packages.
There is no service container, preset stack, registry, reflection discovery,
environment configuration, or global executor.

## Why not add a timeout policy?

A goroutine raced against a timer cannot stop arbitrary synchronous work. The
caller context owns the total deadline, and adapters may derive shorter
scope-specific deadlines.

## Is the budget distributed?

No. The built-in implementation is process-local. A distributed implementation
must explicitly implement `WorkBudget` and document availability, consistency,
latency, and failure semantics.

## Does retry imply idempotency?

No. Callers and protocol adapters remain responsible for replay safety and
idempotency keys.

## Why are observers synchronous?

The core must not own an unbounded telemetry worker lifecycle. A telemetry
adapter can provide a bounded asynchronous queue with explicit shutdown.

## Troubleshooting

### Why does executor construction reject my policy order?

Logical policies must wrap attempt policies. Pass every logical policy before
the first attempt-scoped policy, and inspect each policy's `Descriptor` for its
scope and stable identifier. Duplicate non-repeatable identifiers and invalid
or panicking descriptors are rejected during `NewExecutor`.

### Why does execution remain blocked after the context deadline?

The executor does not move synchronous work to a goroutine. The operation must
observe its context and return; a policy may shorten a deadline but cannot stop
an operation that ignores cancellation. Bound external I/O in the adapter that
owns it.

### Why was a retry or hedge rejected before downstream work?

Check `errors.Is(result.Err, ErrBudgetRejected)` and use `RejectionReasonOf` to
distinguish per-execution, concurrent, rolling-window, resource, lineage, and
duplicate-work limits. A rejection is local admission control and must not be
reported as a downstream failure.

### Why does budget capacity remain occupied?

Every acquired `Permit` must be completed exactly once, and every logical
`WorkBudgetScope` must be closed. `PermitTTL` bounds recovery from abandoned
permits, but expiry is observed only when the budget is used or inspected; the
budget starts no reaper goroutine. Treat `ErrPermitCompleted`,
`ErrPermitExpired`, and `ErrBudgetClosed` as lifecycle errors instead of
discarding them.

### Why is the result timeline empty?

Plain execution intentionally retains no events. Call `WithTimeline` with a
positive bounded capacity, or install an observer with `WithObserver`. Both
return a new executor value and an error, so retain the returned copy only
after validation succeeds.

### Why does a new process start with an empty work budget?

Budget accounting is process-local and is not shared across replicas or
retained across restarts. Roll out gradually and inspect rejection reasons per
process. Use an explicit distributed `WorkBudget` implementation only when the
application requires cluster-wide coordination.
