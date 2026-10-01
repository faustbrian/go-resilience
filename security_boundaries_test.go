package resilience

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func twoScopeConfig(clock Clock) BudgetConfig {
	return BudgetConfig{MaxResources: 1, MaxScopes: 2, MaxAdditionalPerExecution: 2,
		MaxConcurrentAdditional: 1, MaxAdditionalPerWindow: 2,
		AdditionalWindow: time.Minute, PermitTTL: time.Minute, Clock: clock}
}

func TestSecurityScopeCapacityRequiresClose(t *testing.T) {
	for _, settle := range []string{"complete", "expiry"} {
		t.Run(settle, func(t *testing.T) {
			clock := &internalClock{now: time.Unix(1, 0)}
			budget, err := NewBudget(twoScopeConfig(clock))
			if err != nil {
				t.Fatal(err)
			}
			metadata, _ := NewMetadata("logical", "lookup", "shared")
			base, cancel := context.WithCancel(context.Background())
			defer cancel()
			scope, ctx, err := budget.Start(base, metadata)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = scope.Close() })
			other, _, err := budget.Start(context.Background(), metadata)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = other.Close() })
			_, _, permit, err := AdmitAttempt(ctx, OriginOriginal, 0, clock.Now())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = permit.Complete() }()
			cancel()
			if settle == "complete" {
				if err := permit.Complete(); err != nil {
					t.Fatal(err)
				}
			} else {
				clock.now = clock.now.Add(2 * time.Minute)
			}
			third, _, err := budget.Start(context.Background(), metadata)
			if third != nil {
				t.Cleanup(func() { _ = third.Close() })
			}
			if third != nil || RejectionReasonOf(err) != ReasonScopeLimit {
				t.Fatalf("cancellation/%s released an open scope: scope=%v reason=%s", settle, third != nil, RejectionReasonOf(err))
			}
			if err := scope.Close(); err != nil {
				t.Fatal(err)
			}
			replacement, _, err := budget.Start(context.Background(), metadata)
			if err != nil {
				t.Fatalf("Close did not release scope capacity: %v", err)
			}
			t.Cleanup(func() { _ = replacement.Close() })
			if settle == "expiry" && !errors.Is(permit.Complete(), ErrPermitExpired) {
				t.Fatal("expired permit lost expiry classification")
			}
		})
	}
}

func TestSecuritySameResourceScopeAdmission(t *testing.T) {
	budget, err := NewBudget(twoScopeConfig(systemClock{}))
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := NewMetadata("logical", "lookup", "shared")
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		scope, _, err := budget.Start(context.Background(), metadata)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = scope.Close() })
	}
	third, _, err := budget.Start(context.Background(), metadata)
	if third != nil {
		t.Cleanup(func() { _ = third.Close() })
	}
	if !errors.Is(err, ErrBudgetRejected) || third != nil {
		t.Fatalf("third same-resource scope admitted: scope=%v error=%v", third != nil, err)
	}
}

type reentryClock struct {
	budget    *Budget
	scope     *BudgetScope
	active    bool
	held      bool
	reentered bool
}

type scopeLookupContext struct {
	context.Context
	budget   *Budget
	held     bool
	lookedUp bool
}

func (ctx *scopeLookupContext) Value(key any) any {
	if _, ok := key.(budgetContextKey); ok {
		ctx.lookedUp = true
		if ctx.budget.mu.TryLock() {
			ctx.budget.mu.Unlock()
		} else {
			ctx.held = true
		}
	}
	return ctx.Context.Value(key)
}

func TestSecurityContextLookupOutsideAccountingLock(t *testing.T) {
	budget, err := NewBudget(twoScopeConfig(systemClock{}))
	if err != nil {
		t.Fatal(err)
	}
	metadata, _ := NewMetadata("logical", "lookup", "shared")
	scope, attached, err := budget.Start(context.Background(), metadata)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = scope.Close() })
	probe := &scopeLookupContext{Context: attached, budget: budget}
	attempt, _ := NewAttempt(1, OriginOriginal, 0, time.Unix(1, 0))
	permit, err := scope.Acquire(probe, attempt)
	if err != nil {
		t.Fatal(err)
	}
	if err := permit.Complete(); err != nil {
		t.Fatal(err)
	}
	if probe.held || !probe.lookedUp {
		t.Fatalf("Context.Value callback: accounting lock held=%v lookup=%v", probe.held, probe.lookedUp)
	}
	if _, err := scope.Acquire(context.Background(), attempt); !errors.Is(err, ErrBudgetScopeMismatch) {
		t.Fatalf("open mismatched scope error=%v", err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Acquire(context.Background(), attempt); !errors.Is(err, ErrBudgetClosed) {
		t.Fatalf("closed mismatched scope error=%v", err)
	}
}

func (clock *reentryClock) Now() time.Time {
	if clock.scope != nil && !clock.active {
		if !clock.budget.mu.TryLock() {
			clock.held = true
		} else {
			clock.budget.mu.Unlock()
			clock.active = true
			_ = clock.scope.Snapshot()
			clock.active = false
			clock.reentered = true
		}
	}
	return time.Unix(1, 0)
}

func TestSecurityClockReentryDoesNotHoldAccountingLock(t *testing.T) {
	for _, path := range []string{"snapshot", "close", "rejection"} {
		t.Run(path, func(t *testing.T) {
			clock := &reentryClock{}
			budget, err := NewBudget(twoScopeConfig(clock))
			if err != nil {
				t.Fatal(err)
			}
			clock.budget = budget
			metadata, _ := NewMetadata("logical", "lookup", "shared")
			public, ctx, err := budget.Start(context.Background(), metadata)
			if err != nil {
				t.Fatal(err)
			}
			scope := public.(*BudgetScope)
			t.Cleanup(func() { clock.scope = nil; _ = scope.Close() })
			if path == "rejection" {
				attempt, _ := NewAttempt(1, OriginOriginal, 0, clock.Now())
				permit, err := scope.Acquire(ctx, attempt)
				if err != nil {
					t.Fatal(err)
				}
				if err := permit.Complete(); err != nil {
					t.Fatal(err)
				}
			}
			clock.scope = scope
			switch path {
			case "snapshot":
				_ = scope.Snapshot()
			case "close":
				if err := scope.Close(); err != nil {
					t.Fatal(err)
				}
			case "rejection":
				attempt, _ := NewAttempt(1, OriginOriginal, 0, time.Unix(1, 0))
				if _, err := scope.Acquire(ctx, attempt); !errors.Is(err, ErrBudgetRejected) {
					t.Fatal(err)
				}
			}
			if clock.held || !clock.reentered {
				t.Fatalf("clock callback: accounting lock held=%v reentered=%v", clock.held, clock.reentered)
			}
		})
	}
}

func TestSecurityDirectErrorClassification(t *testing.T) {
	attempt, _ := NewAttempt(1, OriginOriginal, 0, time.Unix(1, 0))
	for _, test := range []struct {
		name string
		err  error
		want OutcomeKind
	}{
		{"direct canceled", context.Canceled, OutcomeCancellation},
		{"direct deadline", context.DeadlineExceeded, OutcomeDeadline},
		{"wrapped canceled", fmt.Errorf("operation: %w", context.Canceled), OutcomeOperationFailure},
		{"joined deadline", errors.Join(errors.New("operation"), context.DeadlineExceeded), OutcomeOperationFailure},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := Failure(0, test.err, attempt)
			if result.Err != test.err || result.Outcome.Kind != test.want {
				t.Fatalf("kind=%s want=%s raw error preserved=%v", result.Outcome.Kind, test.want, result.Err == test.err)
			}
		})
	}
	rejection := &BudgetRejectionError{Reason: ReasonExecutionLimit}
	for _, err := range []error{fmt.Errorf("admission: %w", rejection), errors.Join(rejection, errors.New("operation"))} {
		if got := RejectionReasonOf(err); got != "" {
			t.Errorf("wrapped/joined reason=%q want empty", got)
		}
	}
	if got := RejectionReasonOf(rejection); got != ReasonExecutionLimit {
		t.Fatal(got)
	}
}

func TestSecurityMetadataRejectsControlText(t *testing.T) {
	for _, identity := range []string{"logical\nline", "logical\x00id", "logical\xffid"} {
		if _, err := NewMetadata(identity, "lookup", "shared"); !errors.Is(err, ErrInvalidMetadata) {
			t.Fatalf("invalid identity accepted: error=%v", err)
		}
	}
}
