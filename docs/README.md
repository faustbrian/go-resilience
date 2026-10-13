# Documentation

[Repository entry](../README.md) · [Versioned ecosystem](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)

## Getting started

- [Install and quick start](../README.md#install)
- [API](api.md)
- [Executable examples](../example_test.go)
- [Package API](https://pkg.go.dev/github.com/faustbrian/go-resilience/v2)

## Package map

- The [root package](https://pkg.go.dev/github.com/faustbrian/go-resilience/v2)
  uses the `/v2` module suffix. It owns explicit policy composition,
  outcome classification, bounded observation, and shared retry-plus-hedge
  work budgets.
- Historical v1 retains its module identity without a major-version suffix
  and is unsupported.
- There are no nested modules or public subpackages. Focused resilience
  algorithms remain independently versioned companion libraries.

## Concepts and design

- [Budgets](budgets.md)
- [Composition](composition.md)
- [Design](design.md)
- [Errors](errors.md)

## Operations and security

- [Kubernetes](kubernetes.md)
- [Operations](operations.md)
- [Versioned threat model](threat-model.md)
- [Performance](performance.md)
- [Security policy and reporting guidance](../SECURITY.md)
- [Support](../SUPPORT.md)

## Reference and maintenance

- [Source documentation audit](source-documentation-audit.md)
- [Maintainer verification](operations.md#maintainer-verification)
- [FAQ](faq.md)
- [Migration](migration.md)
- [Compatibility policy](../COMPATIBILITY.md)

## Contributing

- [Contribution guide](../CONTRIBUTING.md)
- [Release history](../CHANGELOG.md)
- [License](../LICENSE)

## Complete documentation map

- [API reference](api.md) — `docs/api.md`
- [Benchmark report: 2026-08-02 Darwin arm64](benchmarks/2026-08-02-darwin-arm64.md) — `docs/benchmarks/2026-08-02-darwin-arm64.md`
- [Budget accounting](budgets.md) — `docs/budgets.md`
- [Composition and timing](composition.md) — `docs/composition.md`
- [Design and ownership](design.md) — `docs/design.md`
- [Errors and observation](errors.md) — `docs/errors.md`
- [FAQ](faq.md) — `docs/faq.md`
- [Kubernetes semantics](kubernetes.md) — `docs/kubernetes.md`
- [Migration](migration.md) — `docs/migration.md`
- [Operations and security](operations.md) — `docs/operations.md`
- [Package map](package-map.md) — `docs/package-map.md`
- [Performance](performance.md) — `docs/performance.md`
- [Source documentation audit](source-documentation-audit.md) — `docs/source-documentation-audit.md`
- [Threat model: v2 source contract](threat-model.md) — `docs/threat-model.md`
