# Feature Specification: Gemara publish OCI composite GitHub Action

## Document overview

This specification describes the **composite** GitHub Action shipped from this repository (`action.yml`): **pinned ORAS CLI**, **`oras login`** with caller-provided credentials, publish from an **OCI image layout** (`layout-copy`) or delegation to an **optional `gemara` CLI** (`sdk`), and exposure of the pushed manifest **`digest`** via **`GITHUB_OUTPUT`** (surfaced as the Action output `digest`).

**Key metadata**

- **Action definition:** `action.yml` (composite, single `bash` step)
- **Related design:** [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md), [docs/DESIGN-FOR-REVIEW.md](../../docs/DESIGN-FOR-REVIEW.md)
- **OpenSpec change (tasks / apply workflow):** `openspec/changes/gemara-bundle-publish-github-action/`
- **Normative SHALL blocks:** also tracked under `openspec/changes/gemara-bundle-publish-github-action/specs/gemara-oci-publish-action/spec.md` for OpenSpec tooling; this file is the repo-local, review-friendly spec in the same spirit as [org-infra/specs](https://github.com/complytime/org-infra/tree/main/specs).

**Relationship to org-infra ORAS publishing:** [complytime/org-infra `specs/005-reusable-publish-oras-workflow`](https://github.com/complytime/org-infra/tree/main/specs/005-reusable-publish-oras-workflow) covers a **reusable workflow** that performs multi-file `oras push` plus attestations. **This repository** intentionally stays **transport-only**: it does not assemble Gemara layer manifests or run SBOM/SLSA steps; Pack and OCI contract details live in **go-gemara** (see README).

## Background and motivation

CI callers need a **small, auditable** Action that:

1. Installs a **specific ORAS release** without requiring consumers to maintain separate install steps.
2. Authenticates to the registry using **secrets the workflow supplies** (`oras login`), without echoing tokens.
3. Publishes a bundle that already exists as a valid **OCI image layout** (`oras cp --from-oci-layout`), **or** runs a **pre-installed `gemara` (or compatible) binary** once the SDK exposes a stable push CLI.
4. Emits a **`sha256:…`** (or equivalent) **digest** for downstream signing, promotion, or verification jobs.

## Core user scenarios

### Priority 1: Layout copy to GHCR (default)

A maintainer checks out the repo, produces an OCI layout on disk (e.g. from Pack or another job), then calls this Action with `publish_mode: layout-copy`, `oci_layout_path`, `layout_ref`, `registry`, `repository`, `tag`, and `password` (for example `${{ secrets.GITHUB_TOKEN }}` for GHCR). The Action installs ORAS, logs in, runs `oras cp --from-oci-layout`, and sets `outputs.digest`.

**Test coverage:** `.github/workflows/ci.yml` publishes `testdata/minimal-layout` to a local registry and asserts `steps.publish.outputs.digest` is non-empty.

### Priority 2: SDK mode (optional binary)

Once a **`gemara`** (or compatible) CLI is available on the runner `PATH` or as a file path, the maintainer sets `publish_mode: sdk`, `gemara_binary`, and `sdk_args`. The Action logs in when required, exports `GEMARA_REGISTRY`, `GEMARA_REPOSITORY`, and `GEMARA_TAG`, runs the binary, then resolves the digest the same way as layout mode.

**Test coverage:** optional stub or real CLI job (see [tasks.md](tasks.md)); until stable, **layout-copy** remains the recommended path.

### Priority 3: Local or HTTP registry (CI)

For `localhost` registries, the caller sets `plain_http: true` so the Action skips HTTPS login and uses ORAS plain-HTTP flags consistent with the implementation.

## Edge cases addressed

- **Missing layout:** Fail before copy if `index.json` is absent.
- **HTTPS without password:** Fail with a clear error (unless `plain_http: true`).
- **Digest resolution:** Prefer `oras resolve` on the destination reference; fall back to parsing `oras cp` output when needed.
- **Username default:** When using password auth, default username behavior matches `action.yml` / README (e.g. `GITHUB_ACTOR` when username omitted).

## Functional requirements summary

The Action must:

1. **Pin ORAS** using the `oras_version` input and install the matching official release for the runner OS/architecture.
2. **Login** with `oras login` for HTTPS when `password` is supplied; **skip login** for `plain_http: true` as implemented.
3. **`layout-copy`:** Require `layout_ref` and a valid layout directory; run `oras cp --from-oci-layout` to `registry/repository:tag`.
4. **`sdk`:** Require `gemara_binary`; export `GEMARA_*` env vars; invoke the binary with `sdk_args`.
5. **Output digest:** Append `digest=…` to `GITHUB_OUTPUT` and expose it as the composite output `digest`; fail if digest cannot be determined.

## Scope boundaries

**In scope:** ORAS install pin, registry auth glue, layout copy, optional CLI invocation, digest output.

**Out of scope:** Gemara YAML validation, layer `mediaType` tables, Pack/Unpack implementation, SLSA/SBOM/cosign (caller or org-infra–style workflows), vendoring `gemara` inside this Action.

## Formal requirements (SHALL / scenarios)

### Requirement: Pinned ORAS CLI install

The Action SHALL download and install the ORAS CLI for the runner OS and architecture using the `oras_version` input as the **sole** version selector for the official ORAS release artifact, and SHALL place the `oras` binary on `PATH` for subsequent commands in the same step.

#### Scenario: Default version is used when input omitted

- **WHEN** the workflow invokes the Action without setting `oras_version`
- **THEN** the Action SHALL install the default ORAS version documented in `action.yml` and successfully run `oras version`

#### Scenario: Caller pins a specific ORAS version

- **WHEN** the workflow sets `oras_version` to a supported release (e.g. `1.2.0`)
- **THEN** the Action SHALL install ORAS from the corresponding `oras-project/oras` GitHub release and use that binary for login, copy, and digest resolution

### Requirement: Registry authentication

For HTTPS registries, the Action SHALL run `oras login` against the `registry` input using credentials supplied by the caller, and SHALL NOT print the `password` input to logs.

#### Scenario: HTTPS registry with password

- **WHEN** `plain_http` is not `true` and `password` is non-empty
- **THEN** the Action SHALL pipe `password` to `oras login <registry> -u <username> --password-stdin` and SHALL succeed when credentials are valid

#### Scenario: HTTPS registry without password is rejected

- **WHEN** `plain_http` is not `true` and `password` is empty
- **THEN** the Action SHALL fail with an error indicating `password` is required

#### Scenario: Plain HTTP registry skips login

- **WHEN** `plain_http` is `true`
- **THEN** the Action SHALL skip `oras login` and SHALL still allow publish to proceed with ORAS plain-HTTP flags as implemented

### Requirement: Layout mode publish

In `publish_mode` `layout-copy` (default), the Action SHALL require a valid OCI image layout directory (`index.json` present), require `layout_ref`, and SHALL copy `oci_layout_path:layout_ref` to `registry/repository:tag` using `oras cp --from-oci-layout`.

#### Scenario: Successful layout copy

- **WHEN** the layout path contains `index.json`, `layout_ref` is set, registry destination inputs are valid, and authentication succeeds (or plain HTTP is enabled as required)
- **THEN** `oras cp --from-oci-layout` SHALL exit zero and the artifact SHALL exist at the destination reference

#### Scenario: Missing layout fails fast

- **WHEN** `index.json` is missing under the resolved layout path
- **THEN** the Action SHALL fail before attempting registry copy

### Requirement: SDK mode optional gemara binary

In `publish_mode` `sdk`, the Action SHALL require `gemara_binary` to reference an executable file or a name resolvable on `PATH`, SHALL export `GEMARA_REGISTRY`, `GEMARA_REPOSITORY`, and `GEMARA_TAG` for the invoked process, SHALL run `oras login` when HTTPS rules require it, and SHALL execute the binary with `sdk_args` expanded by the shell.

#### Scenario: SDK mode invokes configured binary

- **WHEN** `publish_mode` is `sdk`, `gemara_binary` points to an existing executable, and registry authentication preconditions are satisfied
- **THEN** the Action SHALL execute that binary with the provided `sdk_args` and a zero exit status SHALL be required to treat publish as successful

#### Scenario: SDK mode without gemara_binary fails

- **WHEN** `publish_mode` is `sdk` and `gemara_binary` is empty
- **THEN** the Action SHALL fail with an error stating `gemara_binary` is required

### Requirement: Digest output

After a successful publish, the Action SHALL write exactly one `digest` line to `GITHUB_OUTPUT` in the form `digest=<algorithm>:<hex>` (e.g. `sha256:…`) representing the manifest digest for `registry/repository:tag`, and SHALL fail if a digest cannot be determined.

#### Scenario: Digest available to downstream steps

- **WHEN** publish completes successfully
- **THEN** the composite output `digest` SHALL equal the value written to `GITHUB_OUTPUT` and SHALL be non-empty

## Success metrics

- **CI green:** `layout-copy` integration publishes a fixture and yields a non-empty `digest`.
- **Consumers** can pin `oras_version` and rely on stable `with:` / `outputs.digest` semantics documented in [README.md](../../README.md).
