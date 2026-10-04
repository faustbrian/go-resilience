# Compatibility Policy

Each releasable directory is an independent Go module and follows semantic
versioning. The root module uses `v<version>` tags; independently releasable
nested modules use `<module-directory>/v<version>`. This repository currently
has no nested modules.

Before `v1`, minor releases MAY contain reviewed breaking changes, but every
break MUST be documented with migration guidance. Patch releases MUST remain
backward compatible. At and after `v1`, incompatible exported API or documented
behavior changes require a new major version.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).

The current source tree defines the v2 contract; public tags and releases,
not the presence of source, establish publication. The
released v1 API snapshot remains immutable in `api/v1.txt`; `api/v2.txt` is
the active source baseline. Historical consumers in go-bulkhead,
go-concurrency-limit, go-fault-injection, go-hedge, go-retry, go-service, and
go-library-tools using v1 retain that distinct identity. V2 publication and
consumer migration evidence are release blockers; sibling `replace`
directives are prohibited. A focused executor's module major alone does not
establish compatibility with Resilience v2: its admission implementation must
consume the v2 context protocol. See [migration](docs/migration.md).
