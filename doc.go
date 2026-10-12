// Package resilience composes explicit synchronous resilience policies and
// owns process-local retry-plus-hedge amplification accounting.
//
// The package deliberately does not implement retry, hedge, circuit-breaker,
// rate-limit, timeout, fallback, bulkhead, semaphore, or cache algorithms.
// Focused packages implement those decisions and adapt them through Policy,
// Stage, Execution, and WorkBudget.
//
// Policies are ordered outer-to-inner. Logical policies must precede attempt
// policies so a logical policy can invoke the attempt stack repeatedly while
// each physical attempt remains independently observable and, with an attached
// scope, budgeted. Policies must preserve that context attachment.
// Execution is synchronous and never moves an operation into a goroutine.
// Cancellation constrains the context passed to work; it cannot preempt an
// operation or policy that ignores that context. Callers own replay safety.
//
// Construct executors and budgets with NewExecutor and NewBudget. Budget scope
// owners must call Close even after cancellation, permit completion or expiry.
// Permit expiry lazily releases accounting, not physical work. Built-in limits
// are process-local and multiply across replicas; they do not limit original
// work like a bulkhead. These boundaries are exercised by the scope-lifecycle
// tests in security_boundaries_test.go and the replica tests in kubernetes_test.go.
package resilience
