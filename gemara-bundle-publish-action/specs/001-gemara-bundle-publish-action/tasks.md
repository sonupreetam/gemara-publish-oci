# Tasks: Gemara Bundle Publish Action

**Input**: Design documents from `specs/001-gemara-bundle-publish-action/`  
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md); optional [research.md](./research.md), [data-model.md](./data-model.md), [contracts/composite-action.md](./contracts/composite-action.md), [quickstart.md](./quickstart.md) — plus **Constitution Check** in `plan.md` per `.specify/memory/constitution.md`

**Tests**: Not mandated as TDD by the spec; verification tasks reference existing CI workflows instead of new unit-test files.

**Organization**: Phases follow user story priority (US1, US2, P1; US3, US4, P2; then polish).

## Format

`- [ ] [TaskID] [P?] [Story?] Description with file path`

---

## Phase 1: Setup

**Purpose**: Confirm Speckit artifact layout and repo branch hygiene.

- [x] T001 Verify directory `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/` contains `spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `tasks.md`, `checklists/requirements.md`, and `contracts/composite-action.md`
- [x] T002 Create git branch `001-gemara-bundle-publish-action` (or equivalent Speckit pattern) from `gemara-bundle-publish-action` repo root so `.specify/scripts/bash/check-prerequisites.sh --json` succeeds when run from `oci-artifact` workspace root

---

## Phase 2: Foundational (blocking)

**Purpose**: Checklist and contract alignment before story sign-off.

**Checkpoint**: Complete before treating user stories as release-ready.

- [x] T003 [P] Reconcile `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/checklists/requirements.md` checkboxes with current `spec.md` and implementation (mark digest/FR-008 items accurately)
- [x] T004 [P] Diff `gemara-bundle-publish-action/action.yml` against `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/contracts/composite-action.md`; update `contracts/composite-action.md` if inputs, outputs, or behavioral bullets drift

---

## Phase 3: User Story 1 — Publish a bundle from CI (Priority: P1)

**Goal**: Declarative CI publishes a root Gemara YAML to a tagged OCI bundle via SDK; invalid input fails before push.

**Independent Test**: Workflow + registry resolve by tag; CI negative path passes.

- [x] T005 [US1] Audit `gemara-bundle-publish-action/tools/publish/main.go` for FR-002 (only `bundle.Assemble`, `bundle.Pack`, `oras.Copy` / remote auth—no parallel OCI layout assembly); fix regressions if found
- [x] T006 [P] [US1] Confirm `.github/workflows/ci.yml` job `publish-invalid-root` meets SC-002 (non-zero exit, no push) using `gemara-bundle-publish-action/tools/publish/testdata/invalid-root.yaml`
- [x] T007 [P] [US1] Document mapping-reference / `fetcher.URI` assembly expectations for consumers in `gemara-bundle-publish-action/README.md` (acceptance scenario 2)
- [x] T008 [US1] Run `go vet` and `go build` in `gemara-bundle-publish-action/tools/publish/` locally or via CI to confirm publish tool builds after any edits

**Checkpoint**: US1 satisfied when CI green and README covers extends/imports via URI.

---

## Phase 4: User Story 2 — Reproducible automation (Priority: P1)

**Goal**: Pinning and SDK correlation documented for audits.

**Independent Test**: Reader can find action ref + SDK pin from repo files.

- [x] T009 [US2] Add or extend `gemara-bundle-publish-action/CHANGELOG.md` with a release-notes template listing **action git ref** and **go-gemara** `require` / `replace` line copied from `gemara-bundle-publish-action/tools/publish/go.mod`
- [x] T010 [P] [US2] Audit `gemara-bundle-publish-action/README.md` sections “Pinning, releases, and reproducibility” and “SDK version pin” for accuracy vs `tools/publish/go.mod`
- [x] T011 [P] [US2] Verify `gemara-bundle-publish-action/examples/workflow-publish-with-pinned-action.yml` uses `id: publish` and documents semver/SHA pinning in comments

**Checkpoint**: US2 satisfied when CHANGELOG or README gives a copy-pasteable SDK pin story.

---

## Phase 5: User Story 3 — Digest output (Priority: P2)

**Goal**: Downstream jobs can read `steps.<id>.outputs.digest` in `algorithm:hex` form.

**Independent Test**: Example workflow echoes digest; E2E asserts non-empty `sha256:` prefix.

- [x] T012 [US3] Verify `gemara-bundle-publish-action/action.yml` exposes `outputs.digest` from `steps.publish.outputs.digest` and `tools/publish/main.go` writes `GITHUB_OUTPUT` digest key after remote `Resolve`
- [x] T013 [P] [US3] Audit `gemara-bundle-publish-action/examples/workflow-publish-with-pinned-action.yml` and `examples/workflow-publish-with-docker-image.yml` for digest consumption examples
- [x] T014 [US3] **E2E / SC-004**: Automated job **`e2e-publish-ghcr`** in `.github/workflows/ci.yml` performs the same publish + digest checks as `e2e-ghcr.yml` on push and same-repo PRs; maintainers may still use `e2e-ghcr.yml` (`workflow_dispatch`). Record first successful run URL under `plan.md` **E2E evidence**.

**Checkpoint**: US3 satisfied when digest appears in logs and optional E2E run documented.

---

## Phase 6: User Story 4 — Split transport (Priority: P2)

**Goal**: Documented two-phase path without pack logic in transport.

**Independent Test**: README links transport-only action; no duplicate pack in this repo.

- [x] T015 [US4] Confirm `gemara-bundle-publish-action/README.md` “Two-phase publish” links `https://github.com/sonupreetam/gemara-publish-oci` and states SDK/transport boundary
- [x] T016 [P] [US4] Confirm `gemara-bundle-publish-action/examples/README.md` references transport-only workflow and E2E layout expectations at high level

**Checkpoint**: US4 satisfied when cross-links exist and remain accurate.

---

## Phase 7: Polish and governance (cross-cutting)

- [x] T017 [P] Update `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/spec.md` **Status** to `Ready` with SC-004 satisfied by automated **`e2e-publish-ghcr`** (and optional `plan.md` run URL after first green CI)
- [x] T018 Append E2E evidence (workflow run URL or date + digest sample) to `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/plan.md` Phase 0/1 footer or linked note
- [x] T019 [P] Add short “Governance / #63” subsection to `gemara-bundle-publish-action/README.md` pointing to [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63) for **maintainers** and **Marketplace** (program tasks outside this spec’s FRs)
- [ ] T020 [P] When `github.com/gemaraproj/go-gemara` publishes a **tagged** release that **includes** `github.com/gemaraproj/go-gemara/bundle`, remove `replace` in `gemara-bundle-publish-action/tools/publish/go.mod`, bump `require`, run `go mod tidy`, and document in `CHANGELOG.md`. **Blocked (2026-04-21)**: `go get …@v0.3.0` and `…@main` (`v0.3.1-0.20260416211637-ea99f6000be6`) do not contain the `bundle` package—`replace` to `github.com/jpower432/go-gemara` remains required until upstream ships bundle APIs on `gemaraproj/go-gemara`.
- [x] T021 [P] **FR-006 audit**: Confirm `gemara-bundle-publish-action/action.yml` and `tools/publish/main.go` never log `password`, `GEMARA_REGISTRY_PASSWORD`, or raw credential-bearing HTTP traces; record result in `gemara-bundle-publish-action/specs/001-gemara-bundle-publish-action/checklists/speckit-analyze-remediation.md` (G4 row)

---

## Dependencies and execution order

### Phase dependencies

1. **Phase 1** → no deps  
2. **Phase 2** → after Phase 1 (paths exist)  
3. **Phases 3–6** → after Phase 2; **US3/US4** can proceed in parallel with **US1/US2** if staffed (different files)  
4. **Phase 7** → after US1–US4 materially complete (**T017** satisfied with **T014** via CI `e2e-publish-ghcr`)

### User story dependencies

- **US1**: No dependency on other stories (core publish).  
- **US2**: Independent (docs/pinning); may proceed parallel to US1.  
- **US3**: Depends on publish path working (US1) for meaningful E2E.  
- **US4**: Independent (documentation only).

### Parallel opportunities

- **T003** and **T004** together  
- **T006**, **T007**, and **T010**, **T011**, **T013**, **T016**, **T019** (distinct files)

---

## Parallel example: User Story 1

```bash
# After Phase 2, contributors can split:
# Task T006 — CI negative fixture / workflow
# Task T007 — README mapping-reference
# (T005 serial — deep read of main.go)
```

---

## Implementation strategy

### MVP (minimum)

1. Complete Phase 1–2  
2. Complete **Phase 3 (US1)** — shippable publish path  
3. Stop and tag a **v0.x** pre-release if desired

### Full feature (spec Ready)

1. MVP + **US2** (traceability docs)  
2. **US3** digest + **T014** E2E (`e2e-publish-ghcr` in CI)  
3. **US4** doc links  
4. **Phase 7** — status flip + governance note + post-merge SDK pin (T020)

---

## Task summary

| Metric | Value |
|--------|-------|
| **Total tasks** | 21 |
| **US1** | T005–T008 (4) |
| **US2** | T009–T011 (3) |
| **US3** | T012–T014 (3) |
| **US4** | T015–T016 (2) |
| **Setup / Foundational / Polish** | T001–T004, T017–T021 (9) |
| **Suggested MVP scope** | Through **T008** (US1) + **T003–T004** |

---

## Extension hooks (post-tasks)

**Optional:** `/speckit.git.commit` — Prompt: “Commit task changes?” — run if you use Speckit git automation.
