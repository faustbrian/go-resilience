package resilience

import (
	"context"
	"time"
)

// MaxIdentityLength bounds UTF-8 identifier length in bytes, not rune count.
const MaxIdentityLength = 128

// Scope identifies how often a policy participates in an execution.
type Scope string

const (
	// ScopeLogical describes a policy around the complete logical execution.
	ScopeLogical Scope = "logical"
	// ScopeAttempt describes a policy around each physical attempt.
	ScopeAttempt Scope = "attempt"
)

// AttemptOrigin identifies why physical work exists.
type AttemptOrigin string

const (
	// OriginOriginal identifies the first caller-requested operation.
	OriginOriginal AttemptOrigin = "original"
	// OriginRetry identifies sequential repeated work after a failure.
	OriginRetry AttemptOrigin = "retry"
	// OriginHedge identifies concurrent duplicate work for tail latency.
	OriginHedge AttemptOrigin = "hedge"
)

// OutcomeKind classifies an execution without conflating local and downstream failures.
type OutcomeKind string

const (
	// OutcomeSuccess identifies a completed operation without an error.
	OutcomeSuccess OutcomeKind = "success"
	// OutcomeOperationFailure identifies an error returned by downstream work.
	OutcomeOperationFailure OutcomeKind = "operation_failure"
	// OutcomeLocalRejection identifies work denied before downstream invocation.
	OutcomeLocalRejection OutcomeKind = "local_rejection"
	// OutcomeCancellation identifies a cancellation-classified result.
	OutcomeCancellation OutcomeKind = "cancellation"
	// OutcomeDeadline identifies a deadline-classified result, not proof of its origin.
	OutcomeDeadline OutcomeKind = "deadline"
	// OutcomeIgnored identifies work intentionally omitted by a policy.
	OutcomeIgnored OutcomeKind = "ignored"
	// OutcomePolicyFailure identifies invalid or failed policy execution.
	OutcomePolicyFailure OutcomeKind = "policy_failure"
)

// Metadata is validated, immutable logical-execution identity for diagnostics
// and budget matching. Its zero value is invalid. Use operational labels, not
// secrets: bounded printable text is not redaction or cardinality control.
type Metadata struct {
	logicalID string
	operation string
	resource  string
}

// NewMetadata requires each identifier to be nonempty valid UTF-8, printable
// according to unicode.IsPrint, and at most MaxIdentityLength bytes. It neither
// trims whitespace nor normalizes Unicode. Invalid input wraps ErrInvalidMetadata.
func NewMetadata(logicalID, operation, resource string) (Metadata, error) {
	identities := []struct {
		field string
		value string
	}{
		{field: "logical_id", value: logicalID},
		{field: "operation", value: operation},
		{field: "resource", value: resource},
	}
	for _, identity := range identities {
		if identity.value == "" {
			return Metadata{}, invalid(ErrInvalidMetadata, identity.field, "must not be blank")
		}
		if bounded(identity.value) != identity.value {
			return Metadata{}, invalid(ErrInvalidMetadata, identity.field, "must be bounded printable text")
		}
	}
	return Metadata{logicalID: logicalID, operation: operation, resource: resource}, nil
}

// LogicalID returns the stable identity shared by all attempts in an execution.
func (metadata Metadata) LogicalID() string { return metadata.logicalID }

// Operation returns the bounded operation identity used by diagnostics.
func (metadata Metadata) Operation() string { return metadata.operation }

// Resource returns the bounded budget resource identity.
func (metadata Metadata) Resource() string { return metadata.resource }

func (metadata Metadata) validate() error {
	_, err := NewMetadata(metadata.logicalID, metadata.operation, metadata.resource)
	return err
}

// Attempt describes one physical invocation within a logical execution.
type Attempt struct {
	// Ordinal is positive and unique within the owning logical scope.
	Ordinal uint64
	// Origin distinguishes original, retry and hedge work.
	Origin AttemptOrigin
	// ParentOrdinal is zero for original work and an earlier ordinal otherwise.
	ParentOrdinal uint64
	// StartedAt is a nonzero caller-supplied attempt timestamp.
	StartedAt time.Time
}

// NewAttempt validates a nonzero timestamp and structural lineage: original work
// has ordinal 1 and no parent; retry or hedge work has an ordinal above 1 and a
// positive earlier parent. It does not prove uniqueness or actual parent
// admission; BudgetScope.Acquire performs those ownership checks.
func NewAttempt(ordinal uint64, origin AttemptOrigin, parent uint64, startedAt time.Time) (Attempt, error) {
	if ordinal == 0 {
		return Attempt{}, invalid(ErrInvalidAttempt, "ordinal", "must be positive")
	}
	if startedAt.IsZero() {
		return Attempt{}, invalid(ErrInvalidAttempt, "started_at", "must not be zero")
	}
	switch origin {
	case OriginOriginal:
		if ordinal != 1 || parent != 0 {
			return Attempt{}, invalid(ErrInvalidAttempt, "origin", "original work must be ordinal one without a parent")
		}
	case OriginRetry, OriginHedge:
		if ordinal == 1 || parent == 0 || parent >= ordinal {
			return Attempt{}, invalid(ErrInvalidAttempt, "parent_ordinal", "additional work requires an earlier parent")
		}
	default:
		return Attempt{}, invalid(ErrInvalidAttempt, "origin", "is unknown")
	}
	return Attempt{Ordinal: ordinal, Origin: origin, ParentOrdinal: parent, StartedAt: startedAt}, nil
}

// Outcome is the stable classification of a completed execution.
type Outcome struct {
	// Kind distinguishes success, operation failure and locally owned outcomes.
	Kind OutcomeKind
	// Attempt is the terminal attempt; early admission failures may leave it zero.
	Attempt Attempt
}

// Result preserves the typed value, original error, outcome and optional events.
type Result[T any] struct {
	// Value is the operation or policy value; reference-valued data is not copied.
	Value T
	// Err preserves the error supplied by work or a policy, without redaction.
	Err error
	// Outcome classifies the result independently of error-chain traversal.
	Outcome Outcome
	// Events is the retained timeline when enabled, truncated at its capacity.
	// Executor-recorded timelines are caller-owned and may omit terminal events
	// when full. Without a recorder, policy-supplied events are not copied or bounded.
	Events []Event
}

func resultFrom[T any](value T, err error, attempt Attempt) Result[T] {
	kind := OutcomeSuccess
	if err != nil {
		switch { //nolint:staticcheck // Keep the direct-identity classification cases explicit, without traversing caller errors.
		case err == context.Canceled: //nolint:errorlint // Wrapped or joined cancellation remains an operation failure; do not invoke caller Is or Unwrap hooks.
			kind = OutcomeCancellation
		case err == context.DeadlineExceeded: //nolint:errorlint // Only this direct standard sentinel qualifies; do not traverse caller errors.
			kind = OutcomeDeadline
		default:
			kind = OutcomeOperationFailure
		}
	}
	return Result[T]{Value: value, Err: err, Outcome: Outcome{Kind: kind, Attempt: attempt}}
}

// Success constructs a successful policy result without validating attempt.
func Success[T any](value T, attempt Attempt) Result[T] {
	return Result[T]{Value: value, Outcome: Outcome{Kind: OutcomeSuccess, Attempt: attempt}}
}

// Failure preserves value and err without invoking error methods. Only direct
// context.Canceled and context.DeadlineExceeded classify as cancellation and
// deadline; wrappers and joined errors classify as operation failure. A nil err
// yields success. The attempt is not validated here.
func Failure[T any](value T, err error, attempt Attempt) Result[T] {
	return resultFrom(value, err, attempt)
}

// LocalRejection constructs local denial with a zero Value and unvalidated
// attempt. Policy and reason are sanitized to bounded printable UTF-8; cause
// remains accessible through the error chain and is not redacted.
func LocalRejection[T any](attempt Attempt, policy PolicyID, reason string, cause error) Result[T] {
	return Result[T]{
		Err:     &LocalRejectionError{Policy: PolicyID(bounded(string(policy))), Reason: bounded(reason), Cause: cause},
		Outcome: Outcome{Kind: OutcomeLocalRejection, Attempt: attempt},
	}
}

// Ignored constructs a deliberate no-work result with a zero Value and
// unvalidated attempt. It sanitizes reason to bounded printable UTF-8.
func Ignored[T any](attempt Attempt, reason string) Result[T] {
	return Result[T]{
		Err:     &IgnoredError{Reason: bounded(reason)},
		Outcome: Outcome{Kind: OutcomeIgnored, Attempt: attempt},
	}
}

// PolicyFailure constructs a policy-owned failure with a zero Value and
// unvalidated attempt. Policy and stage are sanitized to bounded printable UTF-8;
// cause is preserved and accessible through the error chain.
func PolicyFailure[T any](attempt Attempt, policy PolicyID, stage string, cause error) Result[T] {
	return Result[T]{
		Err:     &PolicyExecutionError{Policy: PolicyID(bounded(string(policy))), Stage: bounded(stage), Cause: cause},
		Outcome: Outcome{Kind: OutcomePolicyFailure, Attempt: attempt},
	}
}
