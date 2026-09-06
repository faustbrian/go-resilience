#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

required=(
	README.md CHANGELOG.md COMPATIBILITY.md CONTRIBUTING.md LICENSE SECURITY.md
	SUPPORT.md THIRD_PARTY_LICENSES.md example_test.go docs/README.md docs/api.md
	docs/budgets.md docs/composition.md docs/design.md docs/errors.md docs/faq.md
	docs/kubernetes.md docs/migration.md docs/operations.md docs/performance.md
	docs/benchmarks/2026-08-02-darwin-arm64.md
	docs/benchmarks/raw/2026-08-02-darwin-arm64.txt
	docs/benchmarks/raw/2026-08-02-darwin-arm64-benchstat.txt
)
for path in "${required[@]}"; do
	if [[ ! -s "${path}" ]]; then
		printf 'required documentation is missing or empty: %s\n' "${path}" >&2
		exit 1
	fi
done

if grep -En 'TODO|TBD|FIXME' README.md CHANGELOG.md COMPATIBILITY.md \
	CONTRIBUTING.md SECURITY.md SUPPORT.md docs/*.md; then
	printf 'unfinished documentation marker found\n' >&2
	exit 1
fi

python3 - <<'PY'
from pathlib import Path
from urllib.parse import unquote
import re


def anchors(document: Path) -> set[str]:
    seen: dict[str, int] = {}
    result: set[str] = set()
    fenced = False
    for line in document.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith(("```", "~~~")):
            fenced = not fenced
            continue
        if fenced:
            continue
        match = re.match(r"^#{1,6}\s+(.+?)\s*#*\s*$", line)
        if not match:
            continue
        slug = match.group(1).strip().lower()
        slug = re.sub(r"[^\w\- ]", "", slug)
        slug = re.sub(r"\s+", "-", slug)
        count = seen.get(slug, 0)
        seen[slug] = count + 1
        result.add(slug if count == 0 else f"{slug}-{count}")
    return result


documents = [
    path for path in Path(".").rglob("*.md")
    if ".golib-tooling" not in path.parts and ".verification" not in path.parts
]
anchor_cache = {document.resolve(): anchors(document) for document in documents}
for document in documents:
    prose: list[str] = []
    fenced = False
    for line in document.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith(("```", "~~~")):
            fenced = not fenced
            continue
        if not fenced:
            prose.append(line)
    for raw_target in re.findall(r"\[[^]]*\]\(([^)]+)\)", "\n".join(prose)):
        target = unquote(raw_target.strip().split()[0])
        if target.startswith(("http://", "https://", "mailto:")):
            continue
        relative, _, fragment = target.partition("#")
        linked = (document.parent / relative).resolve() if relative else document.resolve()
        if not linked.exists():
            raise SystemExit(f"broken relative link in {document}: {raw_target}")
        if fragment and linked.suffix.lower() == ".md":
            linked_anchors = anchor_cache.get(linked)
            if linked_anchors is None:
                linked_anchors = anchors(linked)
                anchor_cache[linked] = linked_anchors
            if fragment.lower() not in linked_anchors:
                raise SystemExit(f"broken local anchor in {document}: {raw_target}")

checked = (Path("README.md"), Path("example_test.go"))
ignored_return = re.compile(
    r"^[^\n:=]*,\s*_\s*(?::=|=).*"
    r"(?:NewMetadata|NewExecutor(?:\[[^]]+\])?|NewBudget|NewAttempt|\.Start|\.Acquire)\s*\("
)
ignored_bare_return = re.compile(
    r"^\s*(?:resilience\.(?:NewMetadata|NewExecutor(?:\[[^]]+\])?|NewBudget|NewAttempt)|"
    r"[A-Za-z_][\w.]*\.(?:Start|Acquire))\s*\("
)
ignored_cleanup = re.compile(
    r"(?:_\s*=|defer)\s+[^\n]+\.(?:Complete|Close)\s*\(|"
    r"^\s*[A-Za-z_][\w.]*\.(?:Complete|Close)\s*\([^)]*\)\s*$"
)
ignored_execute = re.compile(
    r"^\s*(?:_\s*=\s*)?[A-Za-z_][\w.]*\.Execute\s*\("
)
execute_assignment = re.compile(
    r"\b([A-Za-z_]\w*)\s*(?::=|=)\s*[^\n]+\.Execute\s*\("
)
for path in checked:
    content = path.read_text(encoding="utf-8")
    for line_number, line in enumerate(content.splitlines(), 1):
        if (
            ignored_return.search(line)
            or ignored_bare_return.search(line)
            or ignored_cleanup.search(line)
            or ignored_execute.search(line)
        ):
            raise SystemExit(f"ignored public error in {path}:{line_number}")
    for match in execute_assignment.finditer(content):
        result = match.group(1)
        if not re.search(rf"\b{re.escape(result)}\.Err\b", content[match.end():]):
            line_number = content.count("\n", 0, match.start()) + 1
            raise SystemExit(f"unchecked Execute result in {path}:{line_number}")

print("documentation links, anchors, and error handling resolve")
PY

go doc github.com/faustbrian/go-resilience >/dev/null

example_count="$(grep -Ec '^func Example[[:alnum:]_]*\(\)' example_test.go)"
output_count="$(grep -Ec '^[[:space:]]*// Output: .+' example_test.go)"
if [[ "${example_count}" -eq 0 || "${example_count}" -ne "${output_count}" ]]; then
	printf 'every executable example must be nonempty and declare output\n' >&2
	exit 1
fi

GOWORK=off go test -mod=readonly . -run '^Example' -count=1
printf 'documentation contract passed\n'
