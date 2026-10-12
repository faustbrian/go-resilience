package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrBudgetRejected identifies local work-amplification denial.
	ErrBudgetRejected = errors.New("resilience: work budget rejected")
	// ErrBudgetClosed identifies use after a logical budget scope closed.
	ErrBudgetClosed = errors.New("resilience: budget scope closed")
	// ErrBudgetScopeMismatch identifies use of a scope for different metadata.
	ErrBudgetScopeMismatch = errors.New("resilience: budget scope mismatch")
	// ErrBudgetAlreadyAttached identifies nested scope creation on a budget context.
	ErrBudgetAlreadyAttached = errors.New("resilience: budget already attached")
	// ErrBudgetScopeRequired identifies work admission without a shared scope.
	ErrBudgetScopeRequired = errors.New("resilience: budget scope required")
	// ErrPermitCompleted identifies duplicate permit completion.
	ErrPermitCompleted = errors.New("resilience: permit already completed")
	// ErrPermitExpired identifies completion after abandoned-capacity recovery.
	ErrPermitExpired = errors.New("resilience: permit expired")
)

const (
	// MaxBudgetResources bounds retained process-local resource identities.
	MaxBudgetResources = 1_000_000
	// MaxBudgetScopes bounds simultaneously retained logical executions.
	MaxBudgetScopes = 1_000_000
	// DefaultMaxBudgetScopes is used when BudgetConfig.MaxScopes is zero.
	DefaultMaxBudgetScopes = 65_536
	// MaxBudgetAttempts bounds each configured attempt-retention dimension.
	MaxBudgetAttempts = 1_000_000
)

// RejectionReason identifies the bounded admission rule that denied work.
type RejectionReason string

const (
	// ReasonExecutionLimit denies work after an execution exhausts its total allowance.
	ReasonExecutionLimit RejectionReason = "execution_limit"
	// ReasonConcurrentLimit denies work while the resource has no concurrent capacity.
	ReasonConcurrentLimit RejectionReason = "concurrent_limit"
	// ReasonWindowLimit denies work after the resource exhausts its rolling allowance.
	ReasonWindowLimit RejectionReason = "window_limit"
	// ReasonResourceLimit denies creation of a new bounded resource identity.
	ReasonResourceLimit RejectionReason = "resource_limit"
	// ReasonScopeLimit denies creation of a new bounded logical scope.
	ReasonScopeLimit RejectionReason = "scope_limit"
	// ReasonDuplicateWork denies a physical attempt already admitted by the scope.
	ReasonDuplicateWork RejectionReason = "duplicate_work"
	// ReasonOriginalRequired denies additional work before the original attempt.
	ReasonOriginalRequired RejectionReason = "original_required"
	// ReasonUnknownParent denies additional work whose parent was never admitted.
	ReasonUnknownParent RejectionReason = "unknown_parent"
)

// BudgetConfig defines finite process-local amplification limits. Original work
// requires a permit but does not consume additional-work quotas. Except for
// MaxScopes, zero limits/durations or a nil Clock are invalid. Finite dimension
// caps are not a memory quota; their products can retain substantial state.
type BudgetConfig struct {
	// MaxResources bounds retained distinct resources in [1, MaxBudgetResources].
	MaxResources int
	// MaxScopes bounds retained logical scopes in [1, MaxBudgetScopes]. Zero
	// selects DefaultMaxBudgetScopes; open scopes require explicit Close.
	MaxScopes int
	// MaxAdditionalPerExecution bounds cumulative retry-plus-hedge admissions
	// per scope in [1, MaxBudgetAttempts], without refunds on completion or expiry.
	MaxAdditionalPerExecution uint64
	// MaxConcurrentAdditional bounds uncompleted, unexpired additional permits
	// per resource in [1, MaxBudgetAttempts], across all its scopes.
	MaxConcurrentAdditional uint64
	// MaxAdditionalPerWindow bounds recent additional admissions per resource
	// in [1, MaxBudgetAttempts], without refunds on completion or expiry.
	MaxAdditionalPerWindow uint64
	// AdditionalWindow is the positive rolling admission-history duration.
	AdditionalWindow time.Duration
	// PermitTTL is a positive duration from admission. Expiry is observed lazily
	// during scope/budget use, releases accounting only and never stops work.
	PermitTTL time.Duration
	// Clock is borrowed for accounting timestamps and sampled outside locks.
	Clock Clock
}

// WorkBudget is the shared admission contract consumed by retry and hedge policies.
type WorkBudget interface {
	// Start creates a logical owner and returns a derived context carrying it.
	// Callers must close the returned scope on every exit path; executors do not
	// close it for them. Custom implementations own their admission semantics.
	Start(context.Context, Metadata) (WorkBudgetScope, context.Context, error)
}

// WorkBudgetScope owns all physical work for one logical execution.
type WorkBudgetScope interface {
	// Acquire admits one physical attempt using a context attached to this owner.
	// The caller must complete each returned permit, including on panic.
	Acquire(context.Context, Attempt) (Permit, error)
	// Snapshot returns accounting state; custom implementations define freshness.
	Snapshot() BudgetSnapshot
	// Matches reports whether this owner belongs to the supplied logical metadata.
	Matches(Metadata) bool
	// Close rejects further admission and settles logical-scope ownership.
	// Built-in scopes defer release until outstanding permits complete or expire.
	Close() error
}

// Permit owns completion of one admitted physical attempt.
type Permit interface {
	// Complete settles ownership once; callers must handle duplicate or expired
	// completion according to the implementation's lifecycle semantics.
	Complete() error
}

// BudgetSnapshot is an immutable bounded accounting view.
type BudgetSnapshot struct {
	// LogicalID identifies this scope's logical execution.
	LogicalID string
	// Resource identifies the shared resource accounting bucket.
	Resource string
	// AdditionalAdmitted counts cumulative additional admissions for this scope.
	AdditionalAdmitted uint64
	// AdditionalActive counts this scope's uncompleted, unexpired additional permits.
	AdditionalActive uint64
	// AdditionalRecent counts rolling-window additional admissions across the resource.
	AdditionalRecent uint64
	// Closed reports explicit Close, not context cancellation or permit expiry.
	Closed bool
}

// BudgetRejectionError reports local denial without classifying it as downstream failure.
type BudgetRejectionError struct {
	// Reason identifies the denying admission rule.
	Reason RejectionReason
	// Snapshot is accounting state captured at rejection, not a live view.
	Snapshot BudgetSnapshot
}

// Error formats the budget category and reason, not the snapshot.
func (err *BudgetRejectionError) Error() string {
	return fmt.Sprintf("%v: %s", ErrBudgetRejected, err.Reason)
}

// Unwrap returns ErrBudgetRejected.
func (err *BudgetRejectionError) Unwrap() error { return ErrBudgetRejected }

// RejectionReasonOf returns Reason only for a nonnil direct *BudgetRejectionError.
// It does not traverse wrapped/joined errors or call caller error methods; nil,
// typed nil and all other errors return the empty value. It does not sanitize a
// caller-constructed Reason.
func RejectionReasonOf(err error) RejectionReason {
	if rejection, ok := err.(*BudgetRejectionError); ok && rejection != nil { //nolint:errorlint // Only direct admission errors qualify; do not invoke caller As or Unwrap hooks.
		return rejection.Reason
	}
	return ""
}

type resourceBudget struct {
	activeAdditional uint64
	recentAdditional []time.Time
	scopes           uint64
}

// Budget owns concurrency-safe process-local accounting shared by retry and
// hedge attempts. Construct it with NewBudget; the zero value is not usable and
// an initialized Budget must not be copied. The borrowed Clock must support
// concurrent calls. No background reaper or distributed coordination is started.
type Budget struct {
	mu        sync.Mutex
	config    BudgetConfig
	resources map[string]*resourceBudget
	scopes    map[*BudgetScope]struct{}
}

// NewBudget copies and validates finite limits, defaults MaxScopes when zero,
// and constructs an empty budget borrowing Clock. Invalid fields wrap
// ErrInvalidComposition in a ConfigurationError. Limits apply per process, not
// across replicas, and do not determine whether work is safe to retry.
func NewBudget(config BudgetConfig) (*Budget, error) {
	if config.MaxScopes == 0 {
		config.MaxScopes = DefaultMaxBudgetScopes
	}
	checks := []struct {
		field string
		valid bool
	}{
		{"max_resources", config.MaxResources > 0 && config.MaxResources <= MaxBudgetResources},
		{"max_scopes", config.MaxScopes > 0 && config.MaxScopes <= MaxBudgetScopes},
		{"max_additional_per_execution", config.MaxAdditionalPerExecution > 0 && config.MaxAdditionalPerExecution <= MaxBudgetAttempts},
		{"max_concurrent_additional", config.MaxConcurrentAdditional > 0 && config.MaxConcurrentAdditional <= MaxBudgetAttempts},
		{"max_additional_per_window", config.MaxAdditionalPerWindow > 0 && config.MaxAdditionalPerWindow <= MaxBudgetAttempts},
		{"additional_window", config.AdditionalWindow > 0},
		{"permit_ttl", config.PermitTTL > 0},
		{"clock", !nilInterface(config.Clock)},
	}
	for _, check := range checks {
		if !check.valid {
			return nil, invalid(ErrInvalidComposition, check.field, "must be within supported bounds")
		}
	}
	return &Budget{config: config, resources: make(map[string]*resourceBudget), scopes: make(map[*BudgetScope]struct{})}, nil
}

type budgetContextKey struct{}
type attemptContextKey struct{}

type budgetExecutionState struct {
	scope WorkBudgetScope
	mu    sync.Mutex
	next  uint64
}

// Start validates a nonnil, uncanceled context and metadata, lazily reclaims
// expired accounting, then creates one scope attached to a derived context.
// Already attached contexts return ErrBudgetAlreadyAttached; occupied scope or
// resource caps return a BudgetRejectionError. Even identical metadata creates
// distinct owners. The caller must Close the scope; cancellation does not do so.
func (budget *Budget) Start(ctx context.Context, metadata Metadata) (WorkBudgetScope, context.Context, error) {
	if ctx == nil {
		return nil, nil, invalid(ErrInvalidMetadata, "context", "must not be nil")
	}
	if _, attached := BudgetScopeFromContext(ctx); attached {
		return nil, nil, ErrBudgetAlreadyAttached
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if err := metadata.validate(); err != nil {
		return nil, nil, err
	}
	now := budget.config.Clock.Now()
	budget.mu.Lock()
	defer budget.mu.Unlock()
	budget.reapLocked(now)
	if len(budget.scopes) >= budget.config.MaxScopes {
		snapshot := BudgetSnapshot{LogicalID: metadata.logicalID, Resource: metadata.resource}
		return nil, nil, &BudgetRejectionError{Reason: ReasonScopeLimit, Snapshot: snapshot}
	}
	resource := budget.resources[metadata.resource]
	if resource == nil {
		if len(budget.resources) >= budget.config.MaxResources {
			snapshot := BudgetSnapshot{LogicalID: metadata.logicalID, Resource: metadata.resource}
			return nil, nil, &BudgetRejectionError{Reason: ReasonResourceLimit, Snapshot: snapshot}
		}
		resource = &resourceBudget{}
		budget.resources[metadata.resource] = resource
	}
	resource.scopes++
	scope := &BudgetScope{
		budget:   budget,
		metadata: metadata,
		resource: resource,
		permits:  make(map[uint64]*WorkPermit),
		ordinals: make(map[uint64]struct{}),
	}
	budget.scopes[scope] = struct{}{}
	return scope, attachBudgetScope(ctx, scope), nil
}

// WithBudgetScope attaches a nonnil custom scope to a nonnil context without
// replacing an existing owner. It does not start, validate metadata for, or close
// that scope, and does not reject an already-canceled context. Each attachment
// has its own ordinal sequencer; derive further contexts from the returned one
// to share sequencing.
func WithBudgetScope(ctx context.Context, scope WorkBudgetScope) (context.Context, error) {
	if ctx == nil {
		return nil, invalid(ErrInvalidMetadata, "context", "must not be nil")
	}
	if nilInterface(scope) {
		return nil, invalid(ErrInvalidComposition, "budget_scope", "must not be nil")
	}
	if _, attached := BudgetScopeFromContext(ctx); attached {
		return nil, ErrBudgetAlreadyAttached
	}
	return attachBudgetScope(ctx, scope), nil
}

func attachBudgetScope(ctx context.Context, scope WorkBudgetScope) context.Context {
	return context.WithValue(ctx, budgetContextKey{}, &budgetExecutionState{scope: scope})
}

// BudgetScopeFromContext returns the explicitly attached logical owner, including
// a closed one. Nil contexts and missing/invalid attachments return nil, false.
// It does not create a scope or change its accounting.
func BudgetScopeFromContext(ctx context.Context) (WorkBudgetScope, bool) {
	if ctx == nil {
		return nil, false
	}
	state, ok := ctx.Value(budgetContextKey{}).(*budgetExecutionState)
	if !ok {
		return nil, false
	}
	if state == nil {
		return nil, false
	}
	if nilInterface(state.scope) {
		return nil, false
	}
	return state.scope, true
}

// AttemptFromContext returns structurally valid physical-attempt metadata attached
// by admission. Nil contexts and missing/invalid metadata return zero, false.
// Presence alone does not prove the associated permit remains active.
func AttemptFromContext(ctx context.Context) (Attempt, bool) {
	if ctx == nil {
		return Attempt{}, false
	}
	attempt, ok := ctx.Value(attemptContextKey{}).(Attempt)
	if !ok {
		return Attempt{}, false
	}
	if _, err := NewAttempt(attempt.Ordinal, attempt.Origin, attempt.ParentOrdinal, attempt.StartedAt); err != nil {
		return Attempt{}, false
	}
	return attempt, true
}

// AdmitAttempt allocates unique physical-attempt lineage, admits it through
// the attached shared budget, and returns a context carrying that attempt.
// Missing attachment returns ErrBudgetScopeRequired. Failed validation/admission
// may consume an ordinal; callers must not assume consecutive admissions. The
// caller owns Complete on the returned permit and Close on the logical scope.
func AdmitAttempt(ctx context.Context, origin AttemptOrigin, parent uint64, startedAt time.Time) (context.Context, Attempt, Permit, error) {
	if ctx == nil {
		return nil, Attempt{}, nil, invalid(ErrInvalidMetadata, "context", "must not be nil")
	}
	state, ok := ctx.Value(budgetContextKey{}).(*budgetExecutionState)
	if !ok {
		return nil, Attempt{}, nil, ErrBudgetScopeRequired
	}
	if state == nil {
		return nil, Attempt{}, nil, ErrBudgetScopeRequired
	}
	if nilInterface(state.scope) {
		return nil, Attempt{}, nil, ErrBudgetScopeRequired
	}
	ordinal, err := reserveOrdinal(state)
	if err != nil {
		return nil, Attempt{}, nil, err
	}
	attempt, err := NewAttempt(ordinal, origin, parent, startedAt)
	if err != nil {
		return nil, Attempt{}, nil, err
	}
	return admitKnownAttempt(ctx, state, attempt)
}

func admitKnownAttempt(ctx context.Context, state *budgetExecutionState, attempt Attempt) (context.Context, Attempt, Permit, error) {
	advanceOrdinal(state, attempt.Ordinal)
	permit, err := state.scope.Acquire(ctx, attempt)
	if err != nil {
		return nil, Attempt{}, nil, err
	}
	return context.WithValue(ctx, attemptContextKey{}, attempt), attempt, permit, nil
}

func reserveOrdinal(state *budgetExecutionState) (uint64, error) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.next == ^uint64(0) {
		return 0, invalid(ErrInvalidAttempt, "ordinal", "exhausted")
	}
	state.next++
	return state.next, nil
}

func advanceOrdinal(state *budgetExecutionState, ordinal uint64) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.next = max(state.next, ordinal)
}

func (budget *Budget) reapLocked(now time.Time) {
	for scope := range budget.scopes {
		scope.reapLocked(now)
	}
	for identity, resource := range budget.resources {
		budget.pruneWindowLocked(resource, now)
		if resource.scopes == 0 && resource.activeAdditional == 0 && len(resource.recentAdditional) == 0 {
			delete(budget.resources, identity)
		}
	}
}

func (budget *Budget) pruneWindowLocked(resource *resourceBudget, now time.Time) {
	cutoff := now.Add(-budget.config.AdditionalWindow)
	first := 0
	for first < len(resource.recentAdditional) && !resource.recentAdditional[first].After(cutoff) {
		first++
	}
	resource.recentAdditional = append(resource.recentAdditional[:0], resource.recentAdditional[first:]...)
}

// BudgetScope is the concurrency-safe retry-plus-hedge accounting owner created
// by Budget.Start. Its zero value is not usable and it must not be copied.
// Explicit Close is required even after cancellation, completion or expiry;
// original work is tracked but excluded from additional-work limits.
type BudgetScope struct {
	budget             *Budget
	metadata           Metadata
	resource           *resourceBudget
	permits            map[uint64]*WorkPermit
	ordinals           map[uint64]struct{}
	additionalAdmitted uint64
	closed             bool
	released           bool
}

// Acquire atomically admits original, retry or hedge work without queueing for
// quota capacity. The context must be nonnil, uncanceled and attached to this
// scope; lineage must be valid and unique, with admitted parents for additional
// work. Original admission must precede additional work. Closed scopes reject
// new admissions.
// It lazily reaps this scope's expired permits and samples Clock outside locks;
// caller context/clock methods can still block. A permit does not start work.
func (scope *BudgetScope) Acquire(ctx context.Context, attempt Attempt) (Permit, error) {
	if ctx == nil {
		return nil, invalid(ErrInvalidMetadata, "context", "must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := NewAttempt(attempt.Ordinal, attempt.Origin, attempt.ParentOrdinal, attempt.StartedAt); err != nil {
		return nil, err
	}
	// Context attachment is immutable; custom Value implementations must not run
	// while the accounting lock is held. Check closed state before mismatch below.
	attached, ok := BudgetScopeFromContext(ctx)
	now := scope.budget.config.Clock.Now()
	scope.budget.mu.Lock()
	defer scope.budget.mu.Unlock()
	scope.reapLocked(now)
	if scope.closed {
		return nil, ErrBudgetClosed
	}
	if !ok || attached != scope {
		return nil, ErrBudgetScopeMismatch
	}
	if _, exists := scope.ordinals[attempt.Ordinal]; exists {
		return nil, scope.rejectionLocked(ReasonDuplicateWork, now)
	}
	if attempt.Origin != OriginOriginal {
		if _, exists := scope.ordinals[1]; !exists {
			return nil, scope.rejectionLocked(ReasonOriginalRequired, now)
		}
		if _, exists := scope.ordinals[attempt.ParentOrdinal]; !exists {
			return nil, scope.rejectionLocked(ReasonUnknownParent, now)
		}
		if scope.additionalAdmitted >= scope.budget.config.MaxAdditionalPerExecution {
			return nil, scope.rejectionLocked(ReasonExecutionLimit, now)
		}
		scope.budget.pruneWindowLocked(scope.resource, now)
		if uint64(len(scope.resource.recentAdditional)) >= scope.budget.config.MaxAdditionalPerWindow {
			return nil, scope.rejectionLocked(ReasonWindowLimit, now)
		}
		if scope.resource.activeAdditional >= scope.budget.config.MaxConcurrentAdditional {
			return nil, scope.rejectionLocked(ReasonConcurrentLimit, now)
		}
	}

	permit := &WorkPermit{scope: scope, ordinal: attempt.Ordinal, additional: attempt.Origin != OriginOriginal, expiresAt: now.Add(scope.budget.config.PermitTTL)}
	permit.state.Store(permitActive)
	scope.permits[attempt.Ordinal] = permit
	scope.ordinals[attempt.Ordinal] = struct{}{}
	if permit.additional {
		scope.additionalAdmitted++
		scope.resource.activeAdditional++
		scope.resource.recentAdditional = append(scope.resource.recentAdditional, now)
	}
	return permit, nil
}

func (scope *BudgetScope) rejectionLocked(reason RejectionReason, now time.Time) error {
	return &BudgetRejectionError{Reason: reason, Snapshot: scope.snapshotLocked(now)}
}

func (scope *BudgetScope) reapLocked(now time.Time) {
	for ordinal, permit := range scope.permits {
		if !now.Before(permit.expiresAt) && permit.state.CompareAndSwap(permitActive, permitExpired) {
			delete(scope.permits, ordinal)
			if permit.additional {
				scope.resource.activeAdditional--
			}
		}
	}
	if scope.closed && len(scope.permits) == 0 {
		scope.releaseLocked()
	}
}

func (scope *BudgetScope) releaseLocked() {
	if scope.released {
		return
	}
	scope.released = true
	scope.resource.scopes--
	delete(scope.budget.scopes, scope)
}

func (scope *BudgetScope) snapshotLocked(now time.Time) BudgetSnapshot {
	scope.budget.pruneWindowLocked(scope.resource, now)
	active := uint64(0)
	for _, permit := range scope.permits {
		if permit.additional {
			active++
		}
	}
	return BudgetSnapshot{
		LogicalID:          scope.metadata.logicalID,
		Resource:           scope.metadata.resource,
		AdditionalAdmitted: scope.additionalAdmitted,
		AdditionalActive:   active,
		AdditionalRecent:   uint64(len(scope.resource.recentAdditional)),
		Closed:             scope.closed,
	}
}

// Snapshot lazily reaps this scope's expired permits, prunes shared window
// history and returns a value copy. AdditionalActive is scope-local;
// AdditionalRecent is resource-wide. Reading the snapshot does not close a scope.
func (scope *BudgetScope) Snapshot() BudgetSnapshot {
	now := scope.budget.config.Clock.Now()
	scope.budget.mu.Lock()
	defer scope.budget.mu.Unlock()
	scope.reapLocked(now)
	return scope.snapshotLocked(now)
}

// Matches reports whether this scope owns the supplied immutable metadata.
func (scope *BudgetScope) Matches(metadata Metadata) bool {
	return scope.metadata == metadata
}

// Close rejects new work, reaps observed expiry and releases the retained scope
// slot once all its permits settle or expire. It does not stop work, complete
// permits, or refund admission history. Resource identity remains while any
// scopes, active additional permits or rolling history retain it. A second Close
// returns ErrBudgetClosed. Owners must call it even on canceled contexts.
func (scope *BudgetScope) Close() error {
	now := scope.budget.config.Clock.Now()
	scope.budget.mu.Lock()
	defer scope.budget.mu.Unlock()
	if scope.closed {
		return ErrBudgetClosed
	}
	scope.closed = true
	scope.reapLocked(now)
	if len(scope.permits) == 0 {
		scope.releaseLocked()
	}
	return nil
}

const (
	permitActive uint32 = iota + 1
	permitCompleted
	permitExpired
)

// WorkPermit owns one built-in admission and supports concurrent completion.
// It must not be copied; its zero value is not usable. TTL expiry settles
// accounting lazily and does not cancel physical work or close its scope.
type WorkPermit struct {
	scope      *BudgetScope
	ordinal    uint64
	additional bool
	expiresAt  time.Time
	state      atomic.Uint32
}

// Complete settles this permit exactly once and releases additional concurrent
// capacity, but neither refunds admissions nor closes an open scope. Duplicate
// completion returns ErrPermitCompleted; previously observed expiry returns
// ErrPermitExpired. Complete itself does not sample Clock or detect TTL expiry.
func (permit *WorkPermit) Complete() error {
	permit.scope.budget.mu.Lock()
	defer permit.scope.budget.mu.Unlock()
	if !permit.state.CompareAndSwap(permitActive, permitCompleted) {
		if permit.state.Load() == permitExpired {
			return ErrPermitExpired
		}
		return ErrPermitCompleted
	}
	delete(permit.scope.permits, permit.ordinal)
	if permit.additional {
		permit.scope.resource.activeAdditional--
	}
	if permit.scope.closed && len(permit.scope.permits) == 0 {
		permit.scope.releaseLocked()
	}
	return nil
}
