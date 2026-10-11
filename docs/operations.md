# Operations and security

Track outcomes by bounded policy and reason, not raw URL, tenant, credential,
request body, or error text. Useful signals include additional admissions,
budget denials, active permits, permit expiry, execution deadlines, operation
failures, and policy failures.

Sustained execution-limit denial means the configured per-call amplification
is containing work. Sustained concurrent or rolling-window denial means the
resource is overloaded or policies are amplifying too aggressively. Permit
expiry indicates abandoned ownership and should be investigated.

Observers execute synchronously after the operation outcome and budget
accounting settle, and must be fast, concurrency-safe, and nonblocking. Their
panics are recovered. A blocked observer delays return to the caller but cannot
consume the operation deadline or alter the settled result. Export to telemetry
through a bounded adapter, not an unbounded goroutine per event.

Resource and logical identifiers must be low-cardinality operational names,
not customer or secret values. Built-in maps, histories, event timelines, and
identifiers are explicitly bounded.

## Residual risk dispositions

| Risk | Disposition | Owner and review condition |
| --- | --- | --- |
| Process-local limits multiply across replicas and reset on restart. | Accepted Medium amplification risk. | The deploying service owner uses a distributed implementation when cluster-wide authority is required and revisits this when replica count changes. |
| A synchronous observer can delay result delivery. | Accepted Medium availability risk; operation outcome and accounting are already settled. | The telemetry owner keeps callbacks bounded and revisits this if observation requires untrusted or blocking I/O. |
| Preserved operation errors expose their own methods to callers. | Accepted Medium caller-boundary risk; v2 never invokes those methods internally. | The operation owner sanitizes untrusted error implementations and revisits this when errors cross a plugin boundary. |
| Configurations near the exported hard caps can intentionally retain substantial process memory. | Accepted Medium configuration availability risk; every dimension remains finite. | The deploying service owner sizes the product of scope, resource, and attempt limits under load and revisits it when workload cardinality changes. |
| Released v1 invokes arbitrary error classifiers and lacks hard policy and scope caps. | Release blocker for hostile callback boundaries; fixed in the v2 source contract. | Maintainers close this only after v2 publication and direct-consumer migration evidence. |

The [versioned threat model](threat-model.md) records the narrower callback,
context, diagnostic, and physical-work boundaries. Permit completion, context
cancellation, and TTL expiry do not close an open logical scope: its owner
must defer `Close`. TTL frees accounting, not a non-cooperative operation.

## Maintainer verification

This standalone repository has no nested modules, required services or test
tags. Start ordinary changes with `go test ./...`; deterministic tests include
fuzz seeds and saved corpus regressions. No sibling checkout or ecosystem goal
ledger is a prerequisite.

The verifier selected by `.golib.yaml` is a checksum-pinned public Tools binary.
`make inventory` validates manifests; `make check` and `make ci` select bounded
local checks. The additional race selection runs once with a two-minute timeout
and observes four existing contracts: concurrent budget admission never exceeds
capacity, observers can reenter without deadlock, scope close cannot free an
active permit, and operation panic releases work while preserving the panic.
It is not a broad stress or leak campaign.

Documentation changes need relevant structure, links and executable examples.
Public API, lifecycle, concurrency and composition changes additionally retain
their affected compatibility, race and direct-consumer integration evidence.
Shared-budget changes must exercise the published Retry and Hedge contracts;
an unaffected companion repository is not an ordinary development prerequisite.

For short active exploration, select exactly one native target:

```bash
GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzBudgetStateMachine$' -fuzztime=2s -parallel=1 -timeout=2m .
```

Long exploration belongs in scheduled assurance. Mutation is optional local or
scheduled work, not a required PR status; a focused local run is
`golib mutation --module . --package .`. Confirmed defects still need repair
and deterministic regression coverage, while equivalent/invalid mutants retain
their reviewed narrow dispositions.

`make check-full` preserves the complete configured full contract: the original
count-20 concurrency/state-machine selection, all four native 1000x fuzz
targets with their twenty-minute timeouts, documentation and benchmarks. Select
it for the supported material risk or full-assurance boundary, not every commit.
Release qualification remains `golib release check --module .` and
`golib release dry-run --module .`.

The manually selected `root_release_qualification` workflow intentionally uses
an immutable source-built verifier and requires its version to be `dev`. That
diagnostic retains complete root-release checks, native evidence and private
workspace cleanup; it is separate from ordinary public-binary verification and
cannot be combined with `release_dry_run`. A bounded local pass does not certify
full assurance, direct-consumer adoption or release publication. Report skipped,
unavailable and failed boundaries explicitly.
