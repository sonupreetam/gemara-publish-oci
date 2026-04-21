# Data model: Gemara Bundle Publish Action

Derived from [spec.md](./spec.md) Key Entities and functional requirements.

## RootGemaraDocument

| Field | Type | Rules |
|-------|------|--------|
| `path` | string (workspace-relative) | Required; resolved under `working_directory`; MUST exist and be readable. |
| `bytes` | opaque (YAML) | Loaded for `DetectType` / optional `gemara.Load` when `validate=true`. |
| `artifact_kind` | enum (from SDK) | Policy, GuidanceCatalog, ControlCatalog, etc.; drives validation branch. |

**Relationships**: Input to **Assembler** with `fetcher.URI` for remote references.

## BundleManifest (SDK)

| Field | Type | Rules |
|-------|------|--------|
| `bundle_version` | string | Default `"1"`; from action input `bundle_version`. |
| `gemara_version` | string | Optional; from action input `gemara_version`; may be empty. |

**Relationships**: Passed with root file into `bundle.Assemble`; output shape owned by SDK.

## GemaraOCIBundle (SDK)

| Field | Type | Rules |
|-------|------|--------|
| `layout` | in-memory OCI store | Produced by `bundle.Pack`; not persisted on runner disk by default. |
| `root_descriptor` | OCI descriptor | Source ref for `oras.Copy` (digest or tagged `gemara-publish/__root__` fallback). |

**Relationships**: Copied to **RegistryReference** as a tagged manifest DAG.

## RegistryReference

| Field | Type | Rules |
|-------|------|--------|
| `registry` | hostname string | HTTPS endpoint for v1 composite (no `http://` first-class support). |
| `repository` | path string | No host; may contain slashes; registry-specific casing rules apply (e.g. GHCR lowercase). |
| `tag` | string | Remote tag applied by `oras.Copy`. |
| `credentials` | username + password/token | MUST NOT be logged (FR-006). |

**Relationships**: Target of publish; post-push **Resolve(tag)** yields **ManifestDigest**.

## ManifestDigest

| Field | Type | Rules |
|-------|------|--------|
| `value` | string | Format `algorithm:hex` (e.g. `sha256:…`); exposed as action output `digest`. |

## GitHubActionInvocation

| Field | Type | Rules |
|-------|------|--------|
| `runner_os` | string | Primary `ubuntu-latest`. |
| `go_version` | string | `1.25.x` per `action.yml`. |
| `composite_step_id` | string | `publish` — exposes `digest` to caller workflows. |

## State transitions (logical)

1. **Validated** (optional) → **Assembled** → **Packed** → **Copied** → **Resolved** → **Success**  
2. Any failure before remote write → **Failed** (no push)  
3. Failure during/after copy → **Failed** (registry may be inconsistent; no atomicity guarantee per spec)
