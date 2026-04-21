# Feature Specification: Gemara Bundle Publish Action

**Feature Branch**: `001-gemara-bundle-publish-action`  
**Created**: 2026-04-21  
**Status**: Ready  
**Input**: Standardize automation for publishing **Gemara OCI bundles** from
repositories into OCI registries (for example GHCR), aligned with upstream goals
in [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60) (OCI
packaging standard), [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)
(bundle publish action), and implementation work in
[go-gemara#62](https://github.com/gemaraproj/go-gemara/pull/62) (SDK bundle
packaging). The **canonical definition** of bundle layout, manifest shape, and
media types **belongs in the Go SDK**; this specification describes what the
**GitHub Action** must deliver to users and how it **depends on** that contract.

**Traceability to [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)**:
This Speckit feature implements the issue’s **“Create a new repo that has a
GitHub Action … leveraging the SDK”** intent: **FR-002** and **User Story 1**
require bundle create-and-push via **Gemara Go SDK bundle APIs** (assemble,
pack, registry copy semantics)—not a reimplementation of pack or media types in
the Action. The **repository** that hosts this Action is the **dedicated repo**
in that checklist item; **final GitHub org/name** is tracked under **Open
Questions** until transfer. **“Identify action maintainers”** and **“Publish to
Marketplace”** remain **#63 program tasks**; they are **referenced** in this spec
(Assumptions, Non-Goals) but are **not** specified as workflow-level requirements
here.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Publish a bundle from CI (Priority: P1)

A project maintainer wants a **declarative CI step** that turns a **root Gemara
YAML** document (and its resolved dependencies) into a **versioned OCI bundle**
in their registry whenever they tag a release or merge to main.

**Why this priority**: This is the core outcome tracked by
[go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63).

**Independent Test**: Given a minimal repository with a valid root file and
registry credentials, a workflow using the action completes successfully and a
client can resolve the pushed **tag** (and optionally **digest**) on the
registry.

**Acceptance Scenarios**:

1. **Given** valid registry credentials and a root Gemara file in the
   workspace, **When** the action runs with required inputs (registry host,
   repository path, tag, root file path, credentials), **Then** the workflow
   completes without error and the registry contains the artifact at the
   requested tag.
2. **Given** the root document references other Gemara content via supported
   resolution rules (for example mapping-reference URLs), **When** the action
   runs, **Then** the published bundle includes that resolved content according
   to the **SDK bundle contract** (not redefined by the action).
3. **Given** invalid or non-loadable root content and validation enabled,
   **When** the action runs, **Then** the step fails with a clear error before any
   push to the registry.

---

### User Story 2 - Reproducible and governable automation (Priority: P1)

A security-conscious organization wants to **pin** the action version (tag or
commit SHA) and understand **which SDK bundle behavior** produced the artifact,
so audits and incident response can correlate a bundle to known software
versions.

**Why this priority**: Without pinning and traceability, registry artifacts are
hard to trust or reproduce.

**Independent Test**: Documentation and packaging allow a consumer to record
**action revision** and **SDK version** used for a given workflow run.

**Acceptance Scenarios**:

1. **Given** a consumer pins the action to a **specific release or commit**,
   **When** they run the same workflow on the same commit of their repo,
   **Then** behavior is documented as stable except where explicitly tied to a
   movable SDK pin (documented in release notes).
2. **Given** the upstream SDK ships a tagged release that defines the bundle
   contract, **When** the action maintainers cut a release, **Then** that release
   documents which **SDK major/minor** it is validated against.

---

### User Story 3 - Provenance: digest after publish (Priority: P2)

A consumer wants the **OCI image index or manifest digest** (for example
`sha256:…`) after publish for SBOMs, admission policies, or immutable references.

**Why this priority**: Digest output is standard for OCI publish steps but is
not always required for first usable drop.

**Independent Test**: A workflow can pass the digest to a later job or print
it in logs from a **named output**.

**Acceptance Scenarios**:

1. **Given** a successful publish, **When** the action completes, **Then** the
   workflow can read a **step output** containing the digest string in a
   standard `algorithm:hex` form where the registry provides it reliably.

---

### User Story 4 - Split transport (optional path) (Priority: P2)

Some teams prefer **two phases**: (1) produce an **OCI image layout** on disk
using tooling/SDK, (2) **copy** that layout to the registry with minimal glue
(aligned with maintainer guidance: SDK defines what is packed; transport stays
thin).

**Why this priority**: Matches ecosystem patterns and reduces risk of encoding
wrong CLI assumptions in the primary action.

**Independent Test**: Documented pattern or companion workflow produces a layout
that passes a **layout-to-registry** copy and yields the same digest family as
direct publish, within the SDK contract.

**Acceptance Scenarios**:

1. **Given** a valid on-disk OCI layout produced by supported upstream tooling,
   **When** a documented **transport-only** path runs, **Then** the artifact is
   available at the target registry reference without the transport step
   redefining layer media types or manifest fields.

---

### Edge Cases

- Registry returns **401/403**: the action fails with an error that indicates
  auth or permission failure without printing the secret.
- **Network interruption** mid-push: the step fails; the registry may or may
  not show a partial upload; documentation states no guarantee of atomic
  publish unless the registry/SDK combination provides it.
- **Very large** dependency graphs: publish may time out; document runner and
  **job-level** timeout guidance (callers SHOULD set `timeout-minutes` on the
  publish job—see README).
- **SDK pre-release**: during development, the action may depend on a
  non-released SDK revision; releases **must** document the migration to
  `require` on official tags.
- **Plain HTTP registry URL**: not supported as a first-class v1 composite
  input (see Assumptions); failures from misconfigured **http://** hosts are
  expected unless the user supplies an **https://** endpoint or an out-of-band
  transport path.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The action MUST accept inputs sufficient to identify **registry
  host**, **repository path** (without host), **tag**, **registry credentials**,
  and **path to the root Gemara document** within the job workspace.
- **FR-002**: The action MUST perform bundle creation and registry upload using
  the **Gemara Go SDK bundle APIs** (assemble, pack, and registry copy semantics)
  as defined by the **merged or explicitly pinned** SDK revision that
  implements [go-gemara#62](https://github.com/gemaraproj/go-gemara/pull/62) or
  its successor; the action MUST NOT redefine **media types**, **manifest JSON
  shape**, or **layer assembly** independently of the SDK.
- **FR-003**: The action MUST support **optional validation** of the root
  document before assembly, defaulting to **on** unless the user opts out.
- **FR-004**: The action MUST allow setting **bundle metadata** fields that the
  SDK exposes for manifests (for example bundle version and optional Gemara
  version label), with defaults documented.
- **FR-005**: The action MUST support configuring a **working directory** for
  path resolution relative to the repository root.
- **FR-006**: On failure, the action MUST exit non-zero and MUST NOT leak
  credentials in logs.
- **FR-007**: Documentation MUST reference upstream issues
  [#60](https://github.com/gemaraproj/go-gemara/issues/60) and
  [#63](https://github.com/gemaraproj/go-gemara/issues/63), state **SDK ownership**
  of the OCI contract, and describe **end-to-end verification** expectations
  against [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) prior to
  merging SDK changes that affect the bundle layout.
- **FR-008**: The action SHOULD expose the published **manifest digest** as an
  output for downstream jobs once reliably obtainable.

### Non-Goals (this specification)

- **Signing** (Sigstore/cosign) and **full provenance attestations** beyond
  digest capture.
- **Defining** new OCI media types or manifest fields outside what the SDK
  ships.
- **Marketplace listing** mechanics (covered by [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)
  checklist but not specified here at workflow level).

### Key Entities

- **Root Gemara document**: Entry YAML file passed to the SDK assembly path.
- **Gemara OCI bundle**: Artifact format produced by the SDK’s pack operation,
  including descriptors and layers the SDK defines.
- **Registry reference**: Host + repository + tag (and optionally digest) for
  the remote artifact.
- **GitHub Action**: Composite (or container) automation invoked from
  `workflow` jobs on GitHub-hosted or self-hosted runners.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A sample workflow documented in this repository (for example
  [`.github/workflows/e2e-ghcr.yml`](../../.github/workflows/e2e-ghcr.yml),
  the **`e2e-publish-ghcr` job** in [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml),
  or the caller examples under `examples/`) **can** complete on `ubuntu-latest` and
  leave a **pullable** artifact at the configured tag on GHCR (or an equivalent
  test registry). **Evidence** for release readiness is a **successful maintainer
  run** of the E2E workflow, a **green `e2e-publish-ghcr` CI job** on the default
  branch or a same-repository pull request, or another **documented** equivalent
  run—not necessarily every push when jobs are skipped (for example **fork pull
  requests**, where the E2E job is skipped because `packages:write` is unavailable).
  **PR CI** still validates build, negative paths, Docker sketch, and (when not a
  fork PR) live GHCR publish and digest output.
- **SC-002**: With validation enabled, an intentionally **invalid** root file
  causes the job to fail in under 3 minutes without a successful push.
- **SC-003**: A maintainer can correlate a successful run to **action version**
  and **documented SDK pin** from repository metadata alone.
- **SC-004**: Before declaring the feature **Ready**, at least one **end-to-end**
  run is recorded against the SDK change that introduces bundle packaging
  ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) or its merged
  equivalent), including **resolve by tag** (and digest when FR-008 is met).
  **Recording**: append the workflow run URL (and optional sample digest) to
  [plan.md](./plan.md) under **E2E evidence** after the first successful
  **`e2e-publish-ghcr`** or **`e2e-ghcr`** run; the CI job is the default path
  once enabled on the default branch.

## Assumptions

- Consumers use **OCI Distribution v2**-compatible registries (GHCR, Quay,
  etc.) with token or password auth supported by the SDK’s remote client.
- **v1 composite action is HTTPS-only**: **plain HTTP** registry endpoints are
  **out of scope** for the first-class action inputs; labs needing HTTP MUST use
  a **TLS-terminating reverse proxy**, a **documented two-phase** layout +
  transport path, or a **fork/wrapper**—not a required built-in `plain_http`
  input in this spec’s v1.
- **Maintainer roster** and **Marketplace** publication will follow
  [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63) and may live
  under a **gemaraproj** org repository name TBD.
- A **transport-only** companion (for example layout + `oras cp`) may exist;
  this spec’s primary user stories focus on **single-step publish from root
  YAML** for developer ergonomics while the SDK contract stabilizes.

## Open Questions

- **Repository home**: Final GitHub org/repo name and transfer timing
  ([go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)).
- **Single action vs two actions**: Whether **transport-only** and
  **pack-and-publish** remain separate repositories or converge after the SDK
  is released.
- **Registry compatibility**: Any registries that reject multi-layer or
  non-image manifests must be documented as unsupported until addressed
  upstream.

## Clarifications

### Session 2026-04-21

- **SDK vs action boundary**: Pack, manifest, and media types are **SDK
  scope**; the action implements **CI orchestration and credentials wiring** and
  must track SDK releases for breaking changes.
- **Jenn / maintainer sequencing**: Prefer **E2E validation against PR #62**
  before merging SDK bundle support, so the action and SDK stay aligned.
- Q: What digest resolution method does the action standardize on? → A:
  **Post-copy remote Resolve by pushed tag** on the same registry client; emit
  **algorithm:hex** (e.g. `sha256:…`). Eventual-consistency retries are
  **out of scope for v1** unless planning documents a concrete registry class
  that requires them.
- Q: Should v1 of the composite action officially support **plain HTTP**
  registries? → A: **No (Option A)**—**HTTPS-only** for the composite action v1;
  document workarounds (TLS front, two-phase transport, or fork); no required
  `plain_http` (or equivalent) input in this spec’s v1.
- Q: Does this Speckit feature follow [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)
  (“Create a new repo that has a GitHub Action … **leveraging the SDK**”)? → A:
  **Yes** for **Action + SDK**: the Action MUST use SDK bundle APIs (**FR-002**).
  **Yes** for **dedicated repo** as the vehicle for that Action; **org/repo
  naming** is **Open Questions** until transfer. **Maintainers** and
  **Marketplace** are **#63 checklist items** deferred to governance (Assumptions
  / Non-Goals), not deep workflow requirements in this spec.
- **FR-008 vs implementation**: The normative text remains **SHOULD** for
  historical spec wording; the shipped composite action **does** expose
  `outputs.digest` via post-copy remote Resolve—treat as **implementation meets
  or exceeds** the SHOULD unless a future spec revision elevates the bar to
  **MUST**.
