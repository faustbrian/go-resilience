package resilience

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

type printablePolicy struct {
	id    PolicyID
	label string
}

func (policy printablePolicy) Descriptor() PolicyDescriptor {
	return PolicyDescriptor{ID: policy.id, Scope: ScopeLogical}
}

func (policy printablePolicy) Wrap(next Stage[int]) Stage[int] {
	return func(ctx context.Context, execution Execution, operation Operation[int]) Result[int] {
		execution.Emit(EventKind(policy.label), PolicyID(policy.label), policy.label)
		return next(ctx, execution, operation)
	}
}

func TestUnicodeNonPrintableBoundaries(t *testing.T) {
	for _, character := range []rune{'\u202e', '\u2028', '\u2029'} {
		t.Run(fmt.Sprintf("U+%04X", character), func(t *testing.T) {
			label := "before" + string(character) + "after"
			t.Run("metadata", func(t *testing.T) {
				for position := range 3 {
					identities := []string{"logical", "lookup", "shared"}
					identities[position] = label
					if _, err := NewMetadata(identities[0], identities[1], identities[2]); !errors.Is(err, ErrInvalidMetadata) {
						t.Errorf("metadata position %d accepted non-printable Unicode", position)
					}
				}
			})
			t.Run("policy", func(t *testing.T) {
				if _, err := NewExecutor[int](printablePolicy{id: PolicyID(label)}); !errors.Is(err, ErrInvalidComposition) {
					t.Fatal("policy accepted non-printable Unicode")
				}
			})
			t.Run("errors", func(t *testing.T) {
				attempt, _ := NewAttempt(1, OriginOriginal, 0, time.Unix(1, 0))
				cause := errors.New("operation cause")
				rejection := LocalRejection[int](attempt, PolicyID(label), label, cause)
				failure := PolicyFailure[int](attempt, PolicyID(label), label, cause)
				ignored := Ignored[int](attempt, label)
				fields := []string{string(rejection.Err.(*LocalRejectionError).Policy), rejection.Err.(*LocalRejectionError).Reason, //nolint:errorlint,forcetypeassert // These factories return the direct owned error types; the oracle inspects their exact fields.
					string(failure.Err.(*PolicyExecutionError).Policy), failure.Err.(*PolicyExecutionError).Stage, ignored.Err.(*IgnoredError).Reason} //nolint:errorlint // Preserve direct factory-type inspection rather than traversing an error chain.
				for _, field := range fields {
					if field != "before_after" {
						t.Errorf("diagnostic field=%q want sanitized replacement", field)
					}
				}
				if !errors.Is(rejection.Err, cause) || !errors.Is(failure.Err, cause) {
					t.Fatal("diagnostic sanitation lost cause identity")
				}
			})
			t.Run("event", func(t *testing.T) {
				event := unicodeEvent(t, label)
				if event.Kind != "before_after" || event.Policy != "before_after" || event.Reason != "before_after" {
					t.Fatalf("event fields not sanitized: kind=%q policy=%q reason=%q", event.Kind, event.Policy, event.Reason)
				}
			})
		})
	}
}

func unicodeEvent(t *testing.T, label string) Event {
	t.Helper()
	executor, err := NewExecutor[int](printablePolicy{id: "policy", label: label})
	if err != nil {
		t.Fatal(err)
	}
	executor, err = executor.WithTimeline(8)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := NewMetadata("logical", "lookup", "shared")
	if err != nil {
		t.Fatal(err)
	}
	result := executor.Execute(context.Background(), metadata, func(context.Context, Attempt) (int, error) { return 7, nil })
	if result.Err != nil || result.Value != 7 || len(result.Events) != 5 {
		t.Fatalf("execution value=%d error=%v events=%d", result.Value, result.Err, len(result.Events))
	}
	return result.Events[1]
}

func TestUnicodePrintableTextAndByteBoundary(t *testing.T) {
	for _, label := range []string{"lookup café", "搜索 支持", "e\u0301 ☃", strings.Repeat("a", 126) + "é"} {
		if _, err := NewMetadata(label, label, label); err != nil {
			t.Fatalf("printable metadata rejected: %v", err)
		}
		if _, err := NewExecutor[int](printablePolicy{id: PolicyID(label)}); err != nil {
			t.Fatalf("printable policy rejected: %v", err)
		}
		event := unicodeEvent(t, label)
		if string(event.Kind) != label || string(event.Policy) != label || event.Reason != label {
			t.Fatal("printable event text changed")
		}
		if got := bounded(label); got != label {
			t.Fatalf("printable diagnostic text changed to %q", got)
		}
	}
	oversized := strings.Repeat("a", 127) + "é"
	if _, err := NewMetadata(oversized, "lookup", "shared"); !errors.Is(err, ErrInvalidMetadata) {
		t.Fatal("129-byte metadata accepted")
	}
	if got := bounded(oversized); got != strings.Repeat("a", 127) || !utf8.ValidString(got) {
		t.Fatalf("UTF-8 truncation boundary=%q", got)
	}
	attempt, _ := NewAttempt(1, OriginOriginal, 0, time.Unix(1, 0))
	cause := errors.New("operation\u202e diagnostic")
	result := Failure(0, cause, attempt)
	if result.Err != cause || result.Err.Error() != "operation\u202e diagnostic" { //nolint:errorlint // The public contract preserves the exact caller error unchanged, not merely an equivalent cause.
		t.Fatal("raw operation error changed")
	}
}
