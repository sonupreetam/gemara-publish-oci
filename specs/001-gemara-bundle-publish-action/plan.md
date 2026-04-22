# Implementation Plan: Gemara Bundle Publish Action

**Branch**: `001-gemara-bundle-publish-action` (Speckit nominal; run `setup-plan.sh` after checking out this branch) | **Date**: 2026-04-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/001-gemara-bundle-publish-action/spec.md`

**Layout**: The composite action and `tools/publish` live at the **repository root** of [sonupreetam/gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci) (this spec path is `specs/001-gemara-bundle-publish-action/` at that root). Older notes that referred to a nested `gemara-bundle-publish-action/` directory are obsolete.

## Summary

Deliver a **composite GitHub Action** that publishes **Gemara OCI bundles** from a **root Gemara YAML** file using **go-gemara** `bundle` APIs (assemble, pack) and **oras-go** registry transport (`oras.Copy`), aligned with [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63), [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60), and [PR #62](https://github.com/gemaraproj/go-gemara/pull/62). Expose **`digest`** after push; keep **HTTPS-only** for v1 per clarifications; document **two-phase** transport (layout + `oras cp` only) as distinct from this repo’s default **bundle** path (see root README).

## Technical Context

**Language/Version**: Go **1.25.x** (runner via `actions/setup-go@v5`); module `tools/publish`  
**Primary Dependencies**: `github.com/gemaraproj/go-gemara` (pinned via `replace` to PR #62 branch until release), `oras.land/oras-go/v2`  
**Storage**: N/A (in-memory OCI store during pack; no server-side persistence in the action)  
**Testing**: `go vet` / `go build` on `tools/publish`; CI negative fixture; optional `workflow_dispatch` E2E to GHCR  
**Target Platform**: GitHub Actions **`ubuntu-latest`** (primary); composite `bash` + `go run`  
**Project Type**: GitHub composite Action + small Go publish helper  
**Performance Goals**: Invalid-root validation + failed job **under 3 minutes** (SC-002); typical publish bounded by runner network and graph size  
**Constraints**: **HTTPS-only** for v1 composite (no `plain_http` input); secrets via `GEMARA_REGISTRY_PASSWORD` only; no ORAS CLI for manifest assembly  
**Scale/Scope**: Single repo action; optional Docker image sketch for cold CI

## Constitution Check

*GATE: Passed. Source: `.specify/memory/constitution.md`.*

- **I. SDK-owned contract**: **Pass** — `tools/publish` calls `bundle.Assemble`, `bundle.Pack`, `oras.Copy`; no custom media types in the action repo.
- **II. Spec-first**: **Pass** — [spec.md](./spec.md) and [checklists/requirements.md](./checklists/requirements.md) track scope; clarifications recorded.
- **III. CI safety and traceability**: **Pass** — password via env; `outputs.digest`; SDK pin documented in `go.mod` / README.
- **IV. Testing discipline**: **Pass** — `.github/workflows/ci.yml` (vet, build, negative); `.github/workflows/e2e-ghcr.yml` for manual E2E.
- **V. Simplicity / transport split**: **Pass** — single-step default; README documents two-phase vs bundle default.

*Post-design re-check: unchanged — Phase 1 artifacts document interfaces only; no new contract logic in the action layer.*

## Project Structure

### Documentation (this feature)

```text
specs/001-gemara-bundle-publish-action/
├── plan.md              # This file
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/           # Phase 1
│   └── composite-action.md
├── spec.md
├── tasks.md
└── checklists/
    ├── requirements.md
    └── speckit-analyze-remediation.md
```

### Source Code (repository root)

```text
action.yml                 # Composite: setup-go + go run tools/publish
tools/publish/
├── go.mod                 # SDK require + replace until PR #62 released
├── go.sum
├── main.go                # Assemble → Pack (memory) → oras.Copy → Resolve digest
└── testdata/
    ├── invalid-root.yaml  # CI negative fixture
    └── minimal-catalog.yaml  # E2E / local smoke
.github/workflows/
├── ci.yml                 # vet, build, docker sketch, publish-invalid-root
└── e2e-ghcr.yml           # workflow_dispatch GHCR publish
examples/                  # Caller workflows, Dockerfile sketch, README
README.md
```

**Structure Decision**: Single composite action repo with **one** Go entrypoint under `tools/publish/` (no duplicate pack implementation). Optional **Docker** path reuses the same binary.

## Complexity Tracking

No constitution violations requiring justification. *(Leave this section empty of
rows unless a future change violates a principle and needs an explicit waiver.)*

## Phase 0 — Research

**Output**: [research.md](./research.md)

Resolved items: SDK pin strategy, digest resolution, HTTPS v1 scope, relationship to transport-only [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci), CI/E2E split.

## Phase 1 — Design

**Outputs**:

- [data-model.md](./data-model.md) — entities and validation rules derived from the spec.
- [contracts/composite-action.md](./contracts/composite-action.md) — consumer-facing inputs/outputs contract (mirrors `action.yml`).
- [quickstart.md](./quickstart.md) — minimal steps to publish from a clone or caller workflow.

**Agent context**: `.cursor/rules/specify-rules.mdc` updated to reference this `plan.md`.

## Phase 2 — Tasks

**Living task list**: [tasks.md](./tasks.md) (from `/speckit.tasks`). Maps work to user stories US1–US4 and governance follow-ups (#63). Update checkboxes there as work completes.

## E2E evidence (SC-004)

**Default path:** CI job **`e2e-publish-ghcr`** in [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) (runs on push and on pull requests from the same repository; skipped on fork PRs). **Alternative:** [`.github/workflows/e2e-ghcr.yml`](../../.github/workflows/e2e-ghcr.yml) via `workflow_dispatch`.

After the first successful run on GitHub, append below:

- **Date**:
- **Workflow run URL**:
- **Sample digest** (`sha256:…`):

*(Fill after the first green `e2e-publish-ghcr` or `e2e-ghcr` run—the job was added 2026-04-21; local clones cannot produce a run URL.)*
