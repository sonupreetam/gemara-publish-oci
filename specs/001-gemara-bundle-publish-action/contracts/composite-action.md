# Contract: Composite action interface

**Source of truth**: [`action.yml`](../../action.yml) in this repository.  
**Consumers**: GitHub Actions workflows (`jobs.*.steps[].uses`).

**Composite structure**: The action runs two steps: (1) `actions/setup-go@v5` with `go-version: 1.25.x`; (2) a `bash` step with **`id: publish`** that runs `go run "${{ github.action_path }}/tools/publish"` with CLI flags mapped from inputs. Only step **publish** produces `GITHUB_OUTPUT` keys consumed by top-level `outputs.digest`.

## Inputs (required unless noted)

| Name | Required | Type | Semantics |
|------|----------|------|-----------|
| `registry` | yes | string | OCI registry host (HTTPS for v1 composite). |
| `repository` | yes | string | Repository path without host; slashes allowed. |
| `tag` | yes | string | Tag applied on the remote artifact. |
| `file` | yes | string | Path to root Gemara YAML, **relative to** `working_directory`. |
| `username` | yes | string | Registry username (e.g. `github.actor` for GHCR). |
| `password` | yes | secret string | Registry token; passed as env `GEMARA_REGISTRY_PASSWORD` to the tool (not echoed). |
| `validate` | no | string (`"true"` / other) | Default `"true"`; when not `true`, passes `-validate=false` to the tool. |
| `bundle_version` | no | string | Default `"1"`; stored in bundle manifest. |
| `gemara_version` | no | string | Optional; stored if non-empty. |
| `working_directory` | no | string | Default `"."`; `cd` before resolving `file`. |

## Outputs

| Name | Semantics |
|------|-----------|
| `digest` | Manifest digest for the pushed `tag` (`algorithm:hex`), after successful copy and remote resolve. |

## Behavioral contract (normative)

1. The step **MUST** run `go run` on `tools/publish` with flags derived from inputs; implementation **MUST** use **go-gemara** `bundle` APIs for pack semantics (**FR-002**).
2. On success, output **`digest`** MUST be set for downstream `steps.<id>.outputs.digest`.
3. On failure, the step **MUST** exit non-zero and **MUST NOT** print `password` or `GEMARA_REGISTRY_PASSWORD`.

## Versioning

Callers **SHOULD** pin `uses:` to a **release tag** or **full commit SHA**. The SDK version used at runtime is defined by **`tools/publish/go.mod`** at that ref.
