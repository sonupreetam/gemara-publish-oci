# Remediation checklist: `/speckit-analyze` (2026)

**Purpose**: Track every finding from the specification analysis pass against `spec.md`, `plan.md`, and `tasks.md`, plus constitution alignment. Update this file when you close items.

**Related**: [requirements.md](./requirements.md) (spec quality), [../plan.md](../plan.md), [../tasks.md](../tasks.md)

## Findings (all issues)

| ID | Category | Severity | Summary | Remediation | Done |
|----|----------|----------|---------|-------------|------|
| **ENV** | Tooling | MEDIUM | Speckit / nested-monorepo paths are obsolete. | [plan.md](../plan.md) **Layout** at repo root. | [x] |
| **I1** | Inconsistency | MEDIUM | Stale plan note about `master` / `setup-plan` failure. | Replaced with stable monorepo + branch guidance in `plan.md`. | [x] |
| **I2** | Inconsistency | LOW | Plan said “Phase 2 — Tasks (not created here)” while `tasks.md` exists. | Plan now links **Phase 2 — Tasks** → `tasks.md`; doc tree lists `tasks.md`. | [x] |
| **U3** | Coverage / wording | MEDIUM | SC-001 read like “every CI push proves GHCR pullable artifact.” | **SC-001** rewritten in `spec.md`: E2E / documented run vs default PR CI scope. | [x] |
| **G4** | Coverage gap | MEDIUM | No explicit task for **FR-006** (no credential leakage). | **T021** in `tasks.md`; audit: `action.yml` passes secret only via `env:`; `main.go` does not print password value (only env var *name* in usage message). | [x] |
| **A5** | Ambiguity | LOW | “Very large graphs” lacked numeric timeout hint. | `spec.md` edge case + `README.md` suggest `timeout-minutes: 15` (adjustable). | [x] |
| **N6** | Normative drift | LOW | FR-008 **SHOULD** vs shipped **digest** output. | **Clarifications** bullet in `spec.md`: implementation meets/exceeds SHOULD. | [x] |
| **D7** | Duplication | LOW | US1 narrative overlaps FR-001/002. | Accepted; no edit required (readability). | [x] |

## Verification notes (G4 / FR-006)

- **`action.yml`**: `password` flows only to `GEMARA_REGISTRY_PASSWORD` env for the publish step; no `echo` of inputs.
- **`tools/publish/main.go`**: Password read from env into variable passed to `auth.StaticCredential`; errors use `%v` on Go errors (review if a registry ever echoed auth in body—unlikely); usage text references env var **name** only when unset.

## Follow-up (`tasks.md`)

- **T014 / T017**: Resolved by automated **`e2e-publish-ghcr`** in `.github/workflows/ci.yml` and **spec Status → Ready** (SC-001 / SC-004 wording updated). **Maintainers:** paste the first successful Actions run URL + sample digest into `plan.md` **E2E evidence** when available.
- **T020**: **Open** until `github.com/gemaraproj/go-gemara` tags a release that includes the `bundle` package (verified 2026-04-21: `v0.3.0` and pseudo-`main` lack `bundle`; `replace` retained).
