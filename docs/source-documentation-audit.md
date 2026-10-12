# Source documentation audit

## Scope and compact plan

Tier A source-contract documentation at main source
`1db46308fac6291354cf754bdacdeadce7ac5abd`. The root module
`github.com/faustbrian/go-resilience/v2` owns the sole public production package
`resilience`; no nested modules, public subpackages or generated variants exist.
The detached task worktree preserves the unrelated original checkout.

Acceptance: inspect every tracked Go file and every public declaration, field,
interface method and canonical package comment; repair evidenced omissions and
misleading wording while keeping executable ASTs, signatures and fixtures
identical; token inspection distinguishes formatter-inserted field separators.
Direct consumers are explicit focused-policy and retry/hedge budget adapters in
the catalog; no consumer source or behavior is changed.

Named risks: caller-owned synchronous cancellation, callback reentry/concurrency,
shared scope/permit lifecycle, finite but potentially large retention, lazy expiry
versus actual physical work, direct-only error classification and caller-owned
sensitive diagnostics. Smallest checks: formatted comments, all-file AST/token
comparison, rendered root Go docs, native local documentation links/markers and
example structure, changed-content secrets and one ordinary complete-diff review.
Runtime, race, active fuzz, benchmark and consumer/release campaigns are not
needed for unchanged executable inputs.

## Reproducible inventory

Run `git ls-files '*.go' 'go.mod'` for tracked source/module inventory;
`modules.json` and `packages.json` enumerate the sole root module/package.
Run `rg -n '^func (Test|Fuzz|Benchmark|Example)' --glob '*.go'` for test entry
points, and inspect `go:build`, `go:generate`, generated headers, deprecation and
unfinished-work markers across every tracked Go file. At the selected source
there are no Go build tags, generated files, generators, deprecations or
unfinished-work markers. All 20 Go files below were read, including both
`resilience` and external `resilience_test` test packages.

| Tracked Go file | Investigated contract and evidence |
| --- | --- |
| [admission_characterization_test.go](../admission_characterization_test.go) | Inclusive caps, typed nil reason, UTF-8 prefix and 64-policy characterization |
| [benchmark_test.go](../benchmark_test.go) | Two benchmarks: composition comparator and whole budget admission lifecycle |
| [boundary_test.go](../boundary_test.go) | Configuration, nil, lineage, hostile error and diagnostic validation partitions |
| [budget.go](../budget.go) | Budget/scope/permit accounting: missing config/interface/snapshot docs; Close/TTL ownership |
| [budget_internal_test.go](../budget_internal_test.go) | Context state, ordinal progression, exact release and callback lock evidence |
| [budget_test.go](../budget_test.go) | Shared retry/hedge, rolling/cardinality/scope bounds, lineage, expiry/close lifecycle |
| [doc.go](../doc.go) | Canonical package: composition and ownership boundaries |
| [errors.go](../errors.go) | Error categories/wrappers: missing fields/methods; cause visibility |
| [example_support_test.go](../example_support_test.go) | Deterministic example clock fixture |
| [example_test.go](../example_test.go) | Two executable examples: executor success and explicit permit/scope cleanup |
| [executor.go](../executor.go) | Policies, callbacks, execution: missing seam/field docs and lifecycle detail |
| [executor_test.go](../executor_test.go) | Order, error identity, contexts, composition caps and repeatability |
| [hardening_test.go](../hardening_test.go) | Concurrency, observer reentry/latency, synchronous work, cancellation; three fuzz targets |
| [kubernetes_test.go](../kubernetes_test.go) | Replica-local independence, cancellation, expiry-only accounting recovery |
| [model_test.go](../model_test.go) | Reference stack/order model and fourth fuzz target |
| [observation.go](../observation.go) | Events/observers: missing fields/method contract; synchronous delivery limits |
| [outcome_test.go](../outcome_test.go) | Constructors, observer panic, operation panic, copy/lineage, total contexts, events |
| [security_boundaries_test.go](../security_boundaries_test.go) | Explicit Close, resource scope caps, lock reentry, direct errors, printable metadata |
| [types.go](../types.go) | Metadata/lineage/outcomes: missing public field docs and direct-sentinel wording |
| [unicode_boundaries_test.go](../unicode_boundaries_test.go) | Unicode printability and byte-bound diagnostic sanitation |

## Missing versus investigated rationale

Missing/underspecified: exported struct fields and interface methods, public error
methods, explicit construction/zero-value requirements, metadata byte/Unicode
limits, direct-sentinel classification without inferring error origin, observer
delivery/ownership, custom-policy composition/concurrency and detailed
scope/permit accounting.
These are clarified from code and the named tests above, not invented design
history. Narrow guide corrections distinguish observed expiry from TTL passage
and recovered observer panic from policy failure. Existing useful comments are
retained or made more precise.

Investigated: package-owned contracts are documented in [design](design.md),
[budgets](budgets.md), [errors](errors.md), [composition](composition.md) and the
[versioned threat model](threat-model.md). The catalog has no external
specification, official conformance corpus or decision register; no external
protocol version or cryptographic/compliance guarantee applies. Retry safety,
distributed coordination, preemption, secret redaction and callback availability
remain caller or companion responsibilities, not guarantees supplied by this
package. No unresolved rationale or runtime repair is part of this pass.
A source-traced follow-up remains: `constrainContext` restores cancellation and
deadline constraints but not caller budget values when a trusted policy proposes an unscoped context.
`terminalStage` then runs unbudgeted work. Source/guide/threat-model contracts now
state that policies must preserve attachment; fail-closed enforcement would be
a separate behavioral change with focused regression/consumer proof, not this
prose pass. No runtime reproduction or repair is claimed here.

All four fuzz targets, their seeds and model helpers were inspected; two
benchmarks include benchmark-only Failsafe-Go v0.9.8 comparison and a full
budget lifecycle. Both executable examples handle public errors and cleanup;
no executable example change is planned. Platform/replica simulations are
ordinary tests, not evidence of a Kubernetes service qualification.

## API, generation and verification boundary

`.golib.yaml` configures an `apidiff` export-data baseline at `api/v2.txt`;
`api/v1.txt` is historical. These are structural API artifacts, not comment
snapshots, and will remain untouched. `scripts/check-docs.sh` renders `go doc`
and also executes unchanged examples; its local link/anchor/error-handling and
example-output structure can be checked without running those examples. No
other repository script generates a prose Go-doc snapshot. Root `go doc -all`
was inspected using catalog Go 1.27.2 with three owned disposable caches.

## Verification and disposition

- Catalog Go 1.27.2 formatted the six changed production Go files and rendered
  `go doc -all .`; the complete rendered reference was inspected.
- A task-owned standard-library AST/scanner check parsed all 20 tracked Go
  files and validated 179 name-led public declarations, fields/interface methods
  and one canonical package comment. Every comment/position-free AST is equal
  to the selected base. Tokens are equal after one explicit formatting-only
  normalization: documenting `IgnoredError.Reason` expands a one-line struct,
  inserting an implicit field-ending semicolon. The initial strict scanner
  identified this separator; no statement, field type or executable AST changed.
- Both structural API baselines, module files and executable examples are
  byte-identical to the base; `.golib.yaml` and native scripts require no prose
  snapshot regeneration.
- The coordinator-supplied current-source CLI ran `golib source-docs check`
  (one module, 20 Go files) and `golib docs check --local`: passed. This is local
  validation, not public release or hosted CI qualification of the candidate CLI.
- The repository-native local Python link/anchor/error-handling section of
  `scripts/check-docs.sh` passed. All inventory rows, two unchanged example output
  declarations, absence of variants/generators/unfinished markers and
  changed-content secret patterns were checked. Pinned Gitleaks v8.30.1 also
  scanned complete changed-file contents with redacted output: no leaks found.
  `git diff --check` passed.
- One ordinary complete-final-diff Tier A inspection found no unresolved
  documentation, ownership, accuracy or scope defect. The source-traced trusted
  policy attachment boundary above remains an explicit behavioral follow-up,
  not a claim of fixed admission enforcement.
- No runtime tests or unchanged examples were executed. Race, active fuzz,
  benchmark, distributed-service, consumer, scanner fleet and release campaigns
  were not run because no executable or supply-chain input changed.
- Task-owned GOCACHE/GOMODCACHE/GOTMPDIR, checker and new execution output were
  removed on both runs; pre-existing tracked mutation evidence was preserved
  byte-for-byte. Parent-owned verifier/scanner binaries were not removed.

This audit does not claim runtime qualification, hosted CI, remote-main delivery
or a public release; coordinator delivery remains separate.

## Objective enforcement adoption

The local `source-docs-check` target is required by ordinary/full checks and CI.
Ordinary shared CI uses `source_documentation: check` with immutable Tools
v2.4.0 and matching authenticated binary/checksum declaration. Separately
selected qualification source identities and gates remain unchanged. Structural
checks do not replace this maintained semantic audit or native docs/examples.
