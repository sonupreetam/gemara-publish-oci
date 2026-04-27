# Design: `gemara-publish-oci` GitHub Action (draft for review)

| Field | Value |
|-------|--------|
| **Status** | Draft — Option 3 standard re-sign promotion path implemented in composite action |
| **Primary reviewers** | Gemara / go-gemara maintainers |
| **Related upstream work** | [go-gemara#60 — Standardize Artifact Packaging and Distribution via OCI](https://github.com/gemaraproj/go-gemara/issues/60) |
| **Repository** | `OWNER/gemara-publish-oci` on GitHub (e.g. under a personal account or **`gemaraproj`** after transfer) |

---

## 1. Purpose

This document describes the current split of responsibilities between:

- **[go-gemara](https://github.com/gemaraproj/go-gemara)** (Pack, manifest, media types, `oras.Copy` semantics), and  
- **`gemara-publish-oci`** (publish orchestration: registry auth, sign/verify, optional promote),

so OCI publishing in CI stays aligned with [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60)
without duplicating layer-level ORAS descriptor ownership.

It is meant for **maintainer review** before wider adoption (e.g. complytime-policies publish pipelines, shared Actions under `gemaraproj`).

---

## 2. Problem statement (from upstream)

Today Gemara artifacts are produced and distributed in **inconsistent** ways across repositories. [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60) proposes **OCI Artifacts** as the standard packaging format, with:

- A clear notion of a **Gemara bundle** in the SDK.  
- **`Pack` / `Unpack`** (or equivalent) in the Go SDK.  
- **Programmatic resolution** of catalogs (including imports) via **OCI URIs**.  
- Optionally, a **standard GitHub Action** to build/publish bundles.

This Action is one candidate for the **last bullet**: it should remain **thin** and defer semantics to the SDK.

---

## 3. Design principles

| Principle | Implication |
|-----------|-------------|
| **SDK is source of truth** | Manifest shape, `artifactType`, layer `mediaType`s, and Pack output are defined and implemented in **go-gemara**, not in this Action. |
| **Transport vs semantics** | Moving bytes and promoting between registries uses ORAS; bundle meaning is still SDK-owned. |
| **No layer assembly in the Action** | The Action does not handcraft layer descriptor tables; publish semantics are delegated to SDK/compatibility tooling. |
| **Pinning** | Callers pin **`@vX.Y.Z`** or commit SHA; ORAS CLI version is an input (`oras_version`). |

### 3.1 Alternatives considered (Options 1 and 2)

These labels align with the **Option 3 thin caller** pattern used in downstream specs (for example
[complytime-policies quickstart](https://github.com/complytime/complytime-policies/blob/main/specs/001-policy-oci-publish/quickstart.md)).

**Option 1: Fully inline caller workflow**

Implement pack, push, sign, and promotion with shell plus CLIs (`oras`, `docker`, `crane`, `cosign`)
directly in each repository workflow, without a reusable composite Action and without org-level
reusable wrappers.

- **Why rejected:** Duplicates transport and trust behavior across repos; difficult to keep aligned
  with [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60) and to audit under one
  pinned contract. Matches the same trade-off called out in caller publish research (inline
  `docker run` / ORAS-only paths as non-thin alternatives).

**Option 2: Chained org-infra reusable workflows**

Use `workflow_call` into separate reusable workflows (staging publish, sign or verify, Quay promote)
from **`complytime/org-infra`** or a public mirror, each pinned at its own commit SHA.

- **Why rejected as the default thin-caller shape:** Fits orgs that centralize promotion and
  attestation policy (see [org-infra reusable publish discussion](https://github.com/complytime/org-infra/issues/172)),
  but callers depend on **multiple** workflow pins and on infra release cadence; forks and demos
  need a coherent mirror set. Option 3 concentrates orchestration in **one** pinned composite Action
  while org-infra chains remain valid where governance explicitly requires them.

**Option 3: Single pinned composite Action (this repository)**

One `uses:` reference to **`gemara-publish-oci`** at a full SHA; encapsulates publish modes, trust
handling, and optional promotion (architecture below).

---

## 4. High-level architecture (Option 3)

```mermaid
flowchart TB
  subgraph sdk [go-gemara SDK]
    Pack[Pack]
    Store[OCI store or layout]
    Copy[oras.Copy to registry]
    Unpack[Unpack consumer]
  end
  subgraph action [gemara-publish-oci Composite]
    Publish[publish_mode entrypoint]
    SignSrc[sign and verify source]
    Promote[promote to Quay]
    Trust[standard re-sign destination trust]
    Out[refs and digest outputs]
  end
  subgraph ci [Caller workflow]
    Secrets[secrets.GITHUB_TOKEN etc]
    Later[sign SBOM promote optional]
  end
  subgraph reg [OCI registry]
    GHCR[GHCR or other]
  end
  Pack --> Store
  Store --> Run
  Publish --> GHCR
  GHCR --> SignSrc
  SignSrc --> Promote
  Promote --> Trust
  Trust --> Out
  Secrets --> Publish
  GHCR --> Unpack
  ci --> action
  Later --> Out
```

Target end state remains SDK-led publish, where `publish_mode: sdk` invokes a stable go-gemara CLI.

Interim compatibility also includes `publish_mode: gemara-file`, which delegates file-based bundle
pack/push while the composite handles trust and promotion orchestration.

---

## 5. Responsibilities (explicit)

| Component | Owns |
|-----------|------|
| **go-gemara** | `Pack` / `Unpack`, bundle definition, manifest and media types, **`oras.Copy`** from packed content to registry (when exposed as API or CLI). |
| **gemara-publish-oci** | Publish source image, sign/verify source, optionally promote to Quay, enforce trust mode, emit source/destination outputs. |
| **Caller workflow** | Checkout, provide inputs/secrets, set release controls and environment gates, and run authoritative cross-registry verification in its own environment. |

---

## 6. Action specification (surface)

Implementation: single composite step in **`action.yml`** (see repository root).

### 6.1 Inputs (summary)

| Input | Role |
|-------|------|
| `publish_mode` | `layout-copy`, `sdk`, or `gemara-file` |
| `registry`, `repository`, `tag` | Target OCI reference (no scheme in `registry`; standard `host` form) |
| `oci_layout_path` / `pack_path` | Root directory of OCI layout (`layout-copy`) |
| `layout_ref` | Reference inside layout for `oras cp PATH:REF` (`layout-copy`, required in that mode) |
| `gemara_binary`, `sdk_args` | Executable and arguments (`sdk` mode) |
| `file`, `validate`, `bundle_version`, `working_directory` | File-based pack/push delegation (`gemara-file`) |
| `username`, `password` | Registry auth (`password` omitted only when `plain_http: true`) |
| `oras_version` | ORAS CLI release (layout path + digest resolution) |
| `sign_source`, `verify_source` | Source trust controls |
| `promote_to_quay`, `quay_*` | Destination promotion and auth controls |
| `trust_mode`, `sign_destination`, `verify_destination` | Destination trust controls (`resign` is the standard path) |
| `allowed_identity_regex`, `cosign_version` | Signature verification and tooling pinning |
| `plain_http` | For HTTP registries (e.g. CI against `localhost:5000`) |

### 6.2 Environment passed to SDK CLI (`sdk` mode)

For interoperability with a future **`gemara`** CLI, the Action sets:

- `GEMARA_REGISTRY` — same as `registry` input  
- `GEMARA_REPOSITORY` — same as `repository` input  
- `GEMARA_TAG` — same as `tag` input  

(Exact CLI flags remain **TBD** in go-gemara; **`sdk_args`** allows callers to pass subcommands until a stable interface exists.)

### 6.3 Outputs

| Output | Meaning |
|--------|---------|
| `digest` / `source_digest` | Source manifest digest |
| `source_ref` | Source image reference with digest |
| `destination_ref`, `destination_digest` | Destination reference/digest after promotion |
| `verified_source`, `verified_destination` | Verification status booleans |
| `trust_mode` | Effective destination trust mode |

---

## 7. Modes (normative behavior)

### 7.1 `layout-copy` (default)

1. Install ORAS CLI (`oras_version`).  
2. Validate `oci_layout_path` contains `index.json`.  
3. `oras login` unless `plain_http` (anonymous HTTP registry).  
4. `oras cp --from-oci-layout "$path:$layout_ref" "$registry/$repository:$tag"`.  
5. Resolve digest and write `GITHUB_OUTPUT`.

### 7.2 `sdk`

1. Install ORAS (for digest resolution after push).  
2. Require `gemara_binary`; resolve on `PATH` if not an absolute executable path.  
3. Set `GEMARA_*` env vars.  
4. `oras login` as above.  
5. Execute `"$gemara_binary" $sdk_args` (caller-defined; **must** match go-gemara when stable).  
6. Resolve digest; write `GITHUB_OUTPUT`.

---

## 8. Non-goals (this Action)

- Gemara YAML **validation** (SDK / separate check).  
- SLSA/SBOM attestation policy ownership (these remain caller/org standards concerns).  
- Defining **canonical `artifactType`** strings — **go-gemara / Gemara** project.  

---

## 9. Questions for review (Gemara / SDK)

1. Should `copy-referrers` remain a supported compatibility mode now that re-sign is the standard path?
2. Should `gemara-file` remain as compatibility mode long-term, or be replaced by stable `sdk` usage?
3. What minimum caller permissions contract should be enforced by policy checks across org repos?

---

## 10. References

- [go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60)  
- [complyctl 001 — research (OCI layout + `oras.Copy`)](https://github.com/complytime/complyctl/blob/main/specs/001-gemara-native-workflow/research.md)  
- [complytime-policies — OCI publish spec](https://github.com/complytime/complytime-policies/blob/main/docs/oci-publish-spec.md)  
- [org-infra#172 — reusable ORAS publish](https://github.com/complytime/org-infra/issues/172)  
- In-repo: [ARCHITECTURE.md](./ARCHITECTURE.md)  

---

*Document version: 1.1-draft. Maintainer edits via PR welcome.*
