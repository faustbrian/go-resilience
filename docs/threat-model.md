# Threat model: v2 source contract

This model applies to the root `github.com/faustbrian/go-resilience/v2`
source API. Public tags and releases establish actual publication; source
presence alone does not. Released v1.1.0 remains historical
and has no global logical-scope cap. Publication, clean public composition,
and affected maintained consumers are separate delivery requirements.

## Authority and observable boundaries

The package has no network, filesystem, environment, process, or autonomous
background-work authority. It synchronously runs explicit caller operations,
policies, clocks, and optional observers. Caller metadata is bounded printable
text, not a place for credentials. No automatic package logging occurs.

The accounting mutex owns shared resource, logical-scope, and permit state;
the attached execution state has a separate mutex for its ordinal sequencer.
`Start` rejects before allocation when `MaxScopes` is full;
zero selects 65,536, and the supported hard ceiling is 1,000,000. Resource
and additional-attempt dimensions also have hard ceilings; near-ceiling
products can still require substantial memory. Original work is not an
additional-work permit limit or a replacement for a bulkhead.

Scope owners must close scopes explicitly. Cancellation, completed permits,
and expired permits do not close open scopes. A closed scope retains its slot
until its permits settle or expire. Exact-once permit completion and lazy
expiry bound accounting, not the lifetime of non-cooperative physical work.
Budget clocks are sampled before the accounting lock on admission, snapshot,
close, and rejection paths. Context scope lookup also runs before the lock;
closed-scope error precedence and admission checks remain under the lock.
Clock re-entry must itself terminate rather than
recursively request another timestamp forever.

V2 classifies only direct context sentinels and direct budget rejections.
Wrapped/joined errors remain preserved but classify as operation failure;
`RejectionReasonOf` does not traverse wrappers. Package-created diagnostic
identifiers are bounded and sanitized, but arbitrary operation errors and
operation panics remain caller-owned. Callers invoking preserved error methods
or exporting those errors must protect their own diagnostic boundary.

V1 and v2 context keys and nominal policy types differ. A focused executor
must explicitly support the attached version; changing a caller import alone
does not migrate retry/hedge admission. Historical v1 routes must stay explicit,
and any later dual-version executor must reject dual attachment without an
implicit priority or split accounting.

## Owned residual dispositions

| Risk | Severity and rationale | Owner, mitigation, review condition |
| --- | --- | --- |
| A custom `context.Context` blocks, panics, or changes its returned attachment. | Medium, accepted trusted-context availability/integrity risk; scope lookup runs outside the accounting lock, but context methods are caller code. | Application integration owner uses standard immutable derived contexts and bounded methods; the go-resilience maintainer must review validation before supporting plugin-supplied contexts. |
| A clock, operation, policy, or synchronous observer blocks or recursively calls itself. | Medium, accepted trusted-callback availability risk; the package cannot preempt caller code. Observer outcome/accounting have already settled. | Application and telemetry owners supply bounded cooperative callbacks and bounded exporters; review when adding untrusted plugins or callback I/O. |
| Permit TTL releases accounting while physical work may still run. | Medium, accepted conditional amplification risk; non-cooperative work is not stopped by expiry. | Service owner uses cooperative deadlines no longer than TTL and an external original/concurrent-work bulkhead; review whenever TTL, deadlines, or cleanup policy changes. |
| Preserved operation errors/panics or secret-valued identifiers leak through application diagnostics. | Medium, accepted application diagnostic-boundary risk; no package-owned logger exports them by default. | Operation/telemetry owners use operational labels and protected error handling; review new exporters or plugin error boundaries. |
| Limits multiply across replicas or retain large finite configuration products. | Medium, accepted deployment/configuration availability risk; limits are process-local, not a memory quota or distributed authority. | Deploying service owner sizes scope/attempt products and replica budgets, closes scopes in every return path, and uses distributed limits when needed; review workload or replica-count changes. |

## Verification and release scope

Focused ordinary regressions cover a two-scope admission ceiling, explicit
close ownership after cancellation/completion/expiry, bounded clock re-entry
on snapshot/close/rejection, context lookup outside the lock with closed versus
mismatch precedence, direct versus wrapped/joined classification, and
printable diagnostics. Existing permit/lineage/model and cancellation tests
cover shared accounting and total-context enforcement. No local test result
is a hosted scanner result or a public-consumer migration claim. The root is
the sole module; no nested module or unrelated companion release is implied.
