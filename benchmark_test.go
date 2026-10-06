package resilience_test

import (
	"context"
	"testing"
	"time"

	"github.com/failsafe-go/failsafe-go"

	"github.com/faustbrian/go-resilience/v2"
)

var benchmarkResult resilience.Result[int]

type passPolicy struct{ id resilience.PolicyID }

func (policy passPolicy) Descriptor() resilience.PolicyDescriptor {
	return resilience.PolicyDescriptor{ID: policy.id, Scope: resilience.ScopeLogical, Repeatable: true}
}

func (passPolicy) Wrap(next resilience.Stage[int]) resilience.Stage[int] { return next }

func BenchmarkExecutionComposition(b *testing.B) {
	ctx := context.Background()
	metadata := metadataFor(b, "benchmark", "resource")
	operation := func(context.Context, resilience.Attempt) (int, error) { return 1, nil }
	direct := func() (int, error) { return 1, nil }
	failsafeExecutor := failsafe.With[int]()
	plain, err := resilience.NewExecutor[int]()
	if err != nil {
		b.Fatal(err)
	}
	composed, err := resilience.NewExecutor[int](passPolicy{id: "outer"}, passPolicy{id: "inner"})
	if err != nil {
		b.Fatal(err)
	}
	observed, err := plain.WithObserver(resilience.ObserverFunc(func(resilience.Event) {}), 8)
	if err != nil {
		b.Fatal(err)
	}

	benchmarks := []struct {
		name string
		run  func()
	}{
		{name: "direct_function", run: func() { _, _ = direct() }},
		{name: "failsafe_no_policy", run: func() { _, _ = failsafeExecutor.Get(direct) }},
		{name: "resilience_no_policy", run: func() { benchmarkResult = plain.Execute(ctx, metadata, operation) }},
		{name: "resilience_two_pass_policies", run: func() { benchmarkResult = composed.Execute(ctx, metadata, operation) }},
		{name: "resilience_observed", run: func() { benchmarkResult = observed.Execute(ctx, metadata, operation) }},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmark.run()
			}
		})
	}
}

func BenchmarkBudgetAdmissionLifecycle(b *testing.B) {
	clock := &manualClock{now: time.Unix(1, 0)}
	config := validBudgetConfig(clock)
	config.MaxResources = 1
	config.MaxScopes = 1
	config.MaxAdditionalPerExecution = 1
	config.MaxConcurrentAdditional = 1
	config.MaxAdditionalPerWindow = 1
	metadata := metadataFor(b, "benchmark", "resource")
	// Each measured iteration owns a finite one-retry budget and closes its
	// scope. Construction, original admission, retry admission/completion and
	// close are included: this is not comparable to the old admission-only cost.
	run := func() {
		budget, err := resilience.NewBudget(config)
		if err != nil {
			b.Fatal(err)
		}
		scope, ctx, err := budget.Start(context.Background(), metadata)
		if err != nil {
			b.Fatal(err)
		}
		defer func() {
			if closeErr := scope.Close(); closeErr != nil {
				b.Fatal(closeErr)
			}
		}()
		admitOriginal(ctx, b, scope, clock.Now())
		attempt, attemptErr := resilience.NewAttempt(2, resilience.OriginRetry, 1, clock.Now())
		if attemptErr != nil {
			b.Fatal(attemptErr)
		}
		permit, acquireErr := scope.Acquire(ctx, attempt)
		if acquireErr != nil {
			b.Fatal(acquireErr)
		}
		if completeErr := permit.Complete(); completeErr != nil {
			b.Fatal(completeErr)
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		run()
	}
}
