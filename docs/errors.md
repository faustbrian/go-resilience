# Errors and observation

Use `errors.Is` for stable categories and `errors.As` for bounded details:

- `ErrInvalidComposition`, `ErrInvalidMetadata`, `ErrInvalidAttempt`, and
  `ErrNilOperation`;
- `ErrLocalRejection`, `ErrIgnored`, and `ErrPolicyFailure`;
- `ErrBudgetRejected`, `ErrBudgetClosed`, `ErrBudgetScopeMismatch`, and
  `ErrBudgetAlreadyAttached`;
- `ErrPermitCompleted` and `ErrPermitExpired`.

`ConfigurationError`, `LocalRejectionError`, `PolicyExecutionError`, and
package-produced `BudgetRejectionError` expose bounded operational fields.
Core wrapper error strings deliberately omit the arbitrary cause; externally
constructed errors or secret-valued identifiers are not a redaction service.
The original cause remains available through `errors.Is` and
`errors.As`.

V2 outcome classification compares only direct `context.Canceled` and
`context.DeadlineExceeded` values. `RejectionReasonOf` accepts only a direct
`*BudgetRejectionError`. These rules prevent arbitrary, cyclic, panicking, or
blocking error methods from executing inside the library. Wrapped operation
errors remain preserved as `Result.Err` but are classified as operation
failures.

Events identify execution, policy, admission, attempt, cancellation, and
completion transitions. Identity and reason strings are printable and bounded
to 128 bytes; diagnostic constructors replace control or malformed bytes, while
configuration constructors reject them.
Events do not retain results or arbitrary errors. A timeline is caller-owned
and bounded; modifying one result timeline cannot affect an executor or later
call. `EventExecutionCanceled` records both caller cancellation and total
deadline expiry, with the corresponding outcome kind as its bounded reason.
Observers receive the bounded timeline after execution settles, so callback
latency cannot alter operation behavior or budget accounting.
