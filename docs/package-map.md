# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-resilience/v2

`resilience` is a small composition foundation for focused resilience libraries. It provides deterministic generic policy composition, typed outcomes, caller-owned total deadlines, bounded observation, and one process-local work budget shared by retry and hedge. It does not hide a default policy stack or implement focused resilience algorithms.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-resilience/v2`](../doc.go) | Use to compose independently selected resilience policies under one caller-owned deadline and shared bounded work budget. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
