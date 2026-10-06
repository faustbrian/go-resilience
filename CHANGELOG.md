# Changelog

## 2.0.1 - 2026-10-06

### Maintenance

- Update the benchmark-only Failsafe-Go comparator to v0.9.8. Resilience's
  production implementation is unchanged. Applications also importing
  Failsafe can select this version through Go's module graph; its expanded
  bulkhead and circuit-breaker builder interfaces may require additional
  methods in custom implementations or mocks.

- Align installation, support, and focused-executor guidance with the
  published v2 module identity, and correct the Retry v2 consumer
  inventory. Consumer adoption remains separate from source publication.

## 2.0.0

Prepared release metadata; public tags and releases establish publication.
Public release and maintained-consumer adoption remain separate requirements
before security closure.

### Security

- Bound executor composition, budget configuration, and retained logical
  scopes so workload cardinality cannot grow package-owned state without a
  hard ceiling. Scope ownership still requires explicit `Close`.
- Classify only direct context and budget errors without invoking arbitrary
  error methods. Wrapped and joined errors retain their causes but classify
  as operation failure; wrapped budget reasons are not extracted.
- Invoke caller-provided budget clocks outside the accounting lock to permit
  callback re-entry, and look up context scope attachment outside that lock.
- Reject control or malformed bytes in configured identities and sanitize
  bounded error and event fields. Operation error values remain caller-owned.

### Changed

- Introduce the `/v2` source contract with hard configuration limits and
  direct-only classification; released v1 remains unchanged. Migrate imports
  to `/v2` and use compatible focused executors for shared-budget composition.
  Publication and direct-consumer migration remain required before security
  closure.

## 1.1.0 - 2026-10-01

### Changed

- Require Go 1.27.0 or newer. Upgrade the toolchain before adopting this
  release; the exported API and runtime contracts remain unchanged.
- Update the benchmark-only Failsafe-Go comparator to v0.9.7. Its upstream
  budget API changes can affect applications that consume Failsafe directly
  when Go selects the newer module version.

- Keep reusable CI and its checked-out tooling on the same v1.8.4 source
  while retaining the checksum-verified v1.4.0 CLI bootstrap.

### Maintenance

- Adopt the checksum-verified `go-library-tools` v1.4.0 CLI and immutable W14
  reusable workflow without changing the resilience API or runtime behavior.

- Publish complete ecosystem family, ownership, construction, lifecycle,
  documentation, and package-selection metadata through the local and CI
  cohesion gate.
- Replace copied repository-local verification tooling with the released
  `go-library-tools` v1.0.5 workflow while preserving the package's strict
  coverage, mutation, fuzz, benchmark, API, and documentation gates.

### Documentation

- Complete the sole-package map, executable error-handling examples,
  troubleshooting guidance, direct private security route, compatibility
  wording, and documentation gate coverage.

- Add canonical v1 installation, stable Go support, lifecycle and ownership,
  project support, and security-reporting guidance.

- Link ecosystem and Resilience family guidance to the immutable v1.4.0
  documentation release.

- Replace obsolete repository links and completed execution artifacts with a
  standalone, human-oriented documentation structure.

## 1.0.0 - 2026-08-26

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-resilience` identity while preserving its documented API and behavior.

### Documentation

- Link the package README to package-owned documentation.

### Added

- Generic immutable outer-to-inner policy composition with explicit logical
  and physical-attempt scopes.
- Typed common outcomes, stable error categories, bounded metadata, attempt
  lineage, timelines, and panic-safe observers.
- Caller-owned total-context enforcement that prevents custom policies from
  extending or detaching deadlines without moving operations to goroutines.
- Process-local retry-plus-hedge work budgets with per-execution, concurrent,
  rolling-window, resource-cardinality, expiry, and exact-completion bounds.
- Context-coordinated physical-attempt admission for focused retry and hedge
  executors, including unique ordinals and nested-attempt reuse.
- Exact statement coverage, mutation, race, fuzz, model, lifecycle, benchmark,
  API, security, documentation, and clean-consumer gate definitions.
