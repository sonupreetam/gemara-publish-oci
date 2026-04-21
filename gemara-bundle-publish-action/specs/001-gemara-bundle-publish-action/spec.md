# Feature Specification: Gemara Bundle Publish Action

**Feature Branch**: `001-gemara-bundle-publish-action`  
**Created**: 2026-04-21  
**Status**: Draft  
**Input**: Standardize automation for publishing **Gemara OCI bundles** from
repositories into OCI registries (for example GHCR), aligned with upstream goals
in [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60) (OCI
packaging standard), [go-gemara#63](https://github.com/gemaraproj/go-gemara/issues/63)
(bundle publish action), and implementation work in
[go-gemara#62](https://github.com/gemaraproj/go-gemara/pull/62) (SDK bundle
packaging). The **canonical definition** of bundle layout, manifest shape, and
media types **belongs in the Go SDK**; this specification describes what the
**GitHub Action** must deliver to users and how it **depends on** that contract.

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
  timeout guidance.
- **SDK pre-release**: during development, the action may depend on a
  non-released SDK revision; releases **must** document the migration to
  `require` on official tags.

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

- **SC-001**: A sample public workflow documented in this repository completes
  on `ubuntu-latest` and leaves a **pullable** artifact at the configured tag on
  GHCR (or equivalent test registry).
- **SC-002**: With validation enabled, an intentionally **invalid** root file
  causes the job to fail in under 3 minutes without a successful push.
- **SC-003**: A maintainer can correlate a successful run to **action version**
  and **documented SDK pin** from repository metadata alone.
- **SC-004**: Before declaring the feature **Ready**, at least one **end-to-end**
  run is recorded against the SDK change that introduces bundle packaging
  ([PR #62](https://github.com/gemaraproj/go-gemara/pull/62) or its merged
  equivalent), including **resolve by tag** (and digest when FR-008 is met).

## Assumptions

- Consumers use **OCI Distribution v2**-compatible registries (GHCR, Quay,
  etc.) with token or password auth supported by the SDK’s remote client.
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
- **Digest resolution**: Exact method when the registry does not immediately
  resolve digest after copy (retry vs parse tool output).
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
