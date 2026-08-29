# Changelog

## Unreleased

### Maintenance

- Replace copied repository-local verification tooling with the released
  `go-library-tools` v1.0.5 workflow while preserving the package's strict
  coverage, mutation, fuzz, benchmark, API, and documentation gates.

### Documentation

- Replace obsolete repository links and completed execution artifacts with a
  standalone, human-oriented documentation structure.

## 1.0.0 - 2026-08-25

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
