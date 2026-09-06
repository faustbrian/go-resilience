# Security policy

## Supported versions

The latest stable v1 release receives security fixes. Older releases and the
`main` branch are unsupported; upgrade before reporting unless the issue is a
regression under active development.

| Version | Supported |
| --- | --- |
| Latest stable v1 release | Yes |
| Older releases | No |
| `main` | No |

## Reporting a vulnerability

Do not disclose a suspected vulnerability in a public issue. Use the
[private vulnerability report](https://github.com/faustbrian/go-resilience/security/advisories/new).
If private reporting is unavailable, ask a maintainer for a private contact
channel without disclosing the vulnerability.

Do not place credentials, customer data, production identifiers, or exploit
payloads in a public report, initial contact request, fixture, event, benchmark,
or mutation report.

## Model

The core performs no network, filesystem, process, environment, unsafe, or
reflection-driven registration work. It does use reflection narrowly to reject
typed nil interfaces during construction. Identifiers, event history,
resources, work totals, concurrency, rolling windows, and permit lifetime are
bounded.

Custom policies, operations, observers, clocks, and distributed budget
implementations are untrusted caller code. Observer panics are recovered and
callbacks run outside budget locks. Callers remain responsible for operation
idempotency, authorization, secret-safe identifiers, and bounded external IO.
