# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected package's docs.
2. Run `go test ./...` or the affected behavior's narrower tests.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Mutation is optional local or scheduled assurance, not a required PR status.
Confirmed behavioral and security defects still require a focused regression
and repair; equivalent/invalid mutants retain reviewed narrow dispositions.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Ordinary standalone verification needs no sibling checkouts or ecosystem
ledgers. It runs deterministic tests, fuzz seeds and saved corpus regressions:

```bash
go test ./...
```

With the checksum-pinned public verifier selected by `.golib.yaml`, `make check`
and `make ci` run bounded local checks, including the named admission,
callback-reentry and permit/scope-cleanup race assertions once with a two-minute
timeout. CI does not build the verifier from source on the ordinary route.

Select one fuzz target for short active exploration, or one package for
optional local mutation:

```bash
GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzBudgetStateMachine$' -fuzztime=2s -parallel=1 -timeout=2m .
golib mutation --module . --package .
```

`make check-full` retains explicit full assurance. See the
[maintainer verification guide](docs/operations.md#maintainer-verification)
for affected compatibility, direct-consumer, full and release boundaries.
Report every unavailable, skipped or failing check; a local pass is not
hosted-CI or public-release qualification.

## Selected Monthly Assurance

[`assurance.yml`](.github/workflows/assurance.yml) runs on main at 07:43 UTC on
the 8th day of each month. It selects one of the 4 declared root-module fuzz
targets using the shared UTC-week selector; it does not run full campaigns or
promise a complete monthly rotation.
Explicit main-only dispatch accepts one exact declared selector, for example
`^FuzzBudgetStateMachine$`.

The selected risks are metadata/attempt round trips, bounded budget state,
policy-stack order and event history, and typed-error/clocked-event semantics.
Existing assertions compare policy order to a reference and require completed
permits to leave no active additional work. This does not certify concurrent
admission, callback reentry, cancellation or composed resilience behavior;
retain affected race, lifecycle, compatibility and direct-consumer checks.

The immutable shared runner bounds active fuzzing to 30 seconds, two workers
and two CPUs, with a five-minute combined native session, 2 GiB memory and
1 MiB output limits inside a ten-minute job. It installs the declared public
CLI without a source build, uses private disposable caches and keeps failures,
timeouts and exhausted budgets visible. Ordinary deterministic tests still run
seeds/corpus; ordinary CI and its Required contract are unchanged. Full/manual
operations and their original budgets remain in `.golib.yaml`.

Treat findings as owner-attributed diagnostics: promptly repair viable defects
with focused regressions; retain reviewed narrow equivalent/invalid dispositions
rather than weakening tests. Review cadence against useful findings and actual
cost; do not interpret a selected pass as full or release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
