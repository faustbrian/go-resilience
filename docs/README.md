# Documentation

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
- Historical v1 retains its unsuffixed module identity and is unsupported.
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

- [FAQ](faq.md)
- [Migration](migration.md)
- [Compatibility policy](../COMPATIBILITY.md)

## Contributing

- [Contribution guide](../CONTRIBUTING.md)
- [Release history](../CHANGELOG.md)
- [License](../LICENSE)
