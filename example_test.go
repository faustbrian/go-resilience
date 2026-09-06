package resilience_test

import (
	"context"
	"fmt"

	"github.com/faustbrian/go-resilience"
)

func ExampleExecutor() {
	metadata, err := resilience.NewMetadata("request-1", "postal.lookup", "postal:FI")
	if err != nil {
		fmt.Println(err)
		return
	}
	executor, err := resilience.NewExecutor[string]()
	if err != nil {
		fmt.Println(err)
		return
	}
	result := executor.Execute(context.Background(), metadata,
		func(_ context.Context, attempt resilience.Attempt) (string, error) {
			return fmt.Sprintf("attempt-%d", attempt.Ordinal), nil
		},
	)
	if result.Err != nil {
		fmt.Println(result.Err)
		return
	}
	fmt.Println(result.Value, result.Outcome.Kind)
	// Output: attempt-1 success
}

func ExampleBudget() {
	clock := fixedExampleClock{}
	budget, err := resilience.NewBudget(resilience.BudgetConfig{
		MaxResources: 16, MaxAdditionalPerExecution: 2,
		MaxConcurrentAdditional: 1, MaxAdditionalPerWindow: 8,
		AdditionalWindow: exampleDuration, PermitTTL: exampleDuration, Clock: clock,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	metadata, err := resilience.NewMetadata("request-1", "postal.lookup", "postal:FI")
	if err != nil {
		fmt.Println(err)
		return
	}
	scope, ctx, err := budget.Start(context.Background(), metadata)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := scope.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	attempt, err := resilience.NewAttempt(1, resilience.OriginOriginal, 0, clock.Now())
	if err != nil {
		fmt.Println(err)
		return
	}
	permit, err := scope.Acquire(ctx, attempt)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := permit.Complete(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(scope.Snapshot().AdditionalAdmitted)
	// Output: 0
}
