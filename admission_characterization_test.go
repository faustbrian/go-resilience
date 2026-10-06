package resilience_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-resilience/v2"
)

func TestRejectionReasonOfTypedNilIsEmpty(t *testing.T) {
	var rejection *resilience.BudgetRejectionError
	if got := resilience.RejectionReasonOf(rejection); got != "" {
		t.Fatalf("typed-nil rejection reason = %q, want empty", got)
	}
}

func TestBudgetConfigurationAdmitsInclusiveCeilings(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*resilience.BudgetConfig)
	}{
		{"resources", func(config *resilience.BudgetConfig) { config.MaxResources = 1_000_000 }},
		{"scopes", func(config *resilience.BudgetConfig) { config.MaxScopes = 1_000_000 }},
		{"execution", func(config *resilience.BudgetConfig) { config.MaxAdditionalPerExecution = 1_000_000 }},
		{"concurrent", func(config *resilience.BudgetConfig) { config.MaxConcurrentAdditional = 1_000_000 }},
		{"window", func(config *resilience.BudgetConfig) { config.MaxAdditionalPerWindow = 1_000_000 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validBudgetConfig(&manualClock{now: time.Unix(1, 0)})
			test.mutate(&config)
			budget, err := resilience.NewBudget(config)
			if err != nil || budget == nil {
				t.Fatalf("inclusive budget ceiling: budget=%v error=%v", budget, err)
			}
		})
	}
}

func TestIgnoredReasonPreservesSanitizedPrefix(t *testing.T) {
	attempt, err := resilience.NewAttempt(1, resilience.OriginOriginal, 0, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, input, want string
	}{
		{"invalid UTF8", "before\xffafter", "before_after"},
		{"control and replacement rune", "\n\ufffd", "_\ufffd"},
		{"no prefix backfill", strings.Repeat("a", 127) + "éz", strings.Repeat("a", 127)},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := resilience.Ignored[int](attempt, test.input)
			var ignored *resilience.IgnoredError
			if !errors.Is(result.Err, resilience.ErrIgnored) || !errors.As(result.Err, &ignored) {
				t.Fatalf("ignored classification = %v", result.Err)
			}
			if ignored.Reason != test.want {
				t.Fatalf("ignored reason = %q, want %q", ignored.Reason, test.want)
			}
		})
	}
}

func TestExecutorAdmitsExactly64LogicalPolicies(t *testing.T) {
	order := make([]string, 0, 128)
	policies := make([]resilience.Policy[string], 0, 64)
	for index := range 64 {
		policies = append(policies, recordingPolicy{
			id:    resilience.PolicyID("logical-" + strconv.Itoa(index)),
			scope: resilience.ScopeLogical, order: &order,
		})
	}
	executor, err := resilience.NewExecutor[string](policies...)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(executor.Policies()); got != 64 {
		t.Fatalf("retained policies = %d, want 64", got)
	}
	metadata, err := resilience.NewMetadata("logical", "lookup", "resource")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	result := executor.Execute(context.Background(), metadata, func(context.Context, resilience.Attempt) (string, error) {
		calls++
		return "result", nil
	})
	if calls != 1 || result.Value != "result" || result.Err != nil || result.Outcome.Kind != resilience.OutcomeSuccess {
		t.Fatalf("execution: calls=%d value=%q error=%v outcome=%v", calls, result.Value, result.Err, result.Outcome.Kind)
	}
	if len(order) != 128 {
		t.Fatalf("policy entry/exit events = %d, want 128", len(order))
	}
}
