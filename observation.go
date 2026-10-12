package resilience

import (
	"sync"
	"time"
)

// EventKind identifies a bounded execution lifecycle event.
type EventKind string

const (
	// EventExecutionStarted identifies logical execution admission.
	EventExecutionStarted EventKind = "execution_started"
	// EventPolicyEntered identifies entry into one composed policy.
	EventPolicyEntered EventKind = "policy_entered"
	// EventWorkAdmitted identifies budget admission of physical work.
	EventWorkAdmitted EventKind = "work_admitted"
	// EventWorkRejected identifies local budget rejection.
	EventWorkRejected EventKind = "work_rejected"
	// EventAttemptStarted identifies invocation of physical work.
	EventAttemptStarted EventKind = "attempt_started"
	// EventAttemptCompleted identifies return from physical work.
	EventAttemptCompleted EventKind = "attempt_completed"
	// EventExecutionCanceled identifies caller cancellation of an execution.
	EventExecutionCanceled EventKind = "execution_canceled"
	// EventExecutionCompleted identifies the terminal logical outcome.
	EventExecutionCompleted EventKind = "execution_completed"
)

// Event is diagnostic metadata without operation values or errors. Package-emitted
// identity strings are bounded printable UTF-8, not redacted secrets. Callers
// constructing Event literals are responsible for their fields.
type Event struct {
	// Kind identifies the transition, including caller-defined policy events.
	Kind EventKind
	// Policy identifies the owning policy, or is empty for executor transitions.
	Policy PolicyID
	// Reason carries bounded operational detail, not arbitrary error text.
	Reason string
	// LogicalID is the shared execution identity.
	LogicalID string
	// Attempt is the stage's physical-attempt metadata.
	Attempt Attempt
	// At is sampled from the executor's clock at emission, not delivery.
	At time.Time
}

// Observer receives the retained timeline synchronously after execution settles,
// outside internal locks. Implementations must be bounded and concurrency-safe
// for concurrent executions; a slow observer delays result delivery.
type Observer interface {
	// Observe receives one retained event. Execute recovers each callback panic
	// and continues notification; direct calls do not have that protection.
	Observe(Event)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(Event)

// Observe invokes the adapted function directly, without panic recovery.
func (function ObserverFunc) Observe(event Event) { function(event) }

type recorder struct {
	mu       sync.Mutex
	max      int
	clock    Clock
	observer Observer
	events   []Event
}

func (recorder *recorder) emit(event Event) {
	event.At = recorder.clock.Now()
	recorder.mu.Lock()
	if len(recorder.events) < recorder.max {
		recorder.events = append(recorder.events, event)
	}
	recorder.mu.Unlock()
}

func (recorder *recorder) snapshot() []Event {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]Event(nil), recorder.events...)
}

func (recorder *recorder) notify(events []Event) {
	if recorder.observer == nil {
		return
	}
	for _, event := range events {
		func() {
			defer func() { _ = recover() }()
			recorder.observer.Observe(event)
		}()
	}
}
