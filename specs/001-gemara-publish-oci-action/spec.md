# Feature Specification: Gemara publish OCI composite GitHub Action

## Document overview

This specification describes the **composite** GitHub Action shipped from this repository
(`action.yml`): publish from an OCI layout, optional SDK CLI invocation, or file-based bundle
compatibility mode, plus keyless sign/verify and optional GHCR -> Quay promotion with explicit
trust modes.

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
4. Emits stable outputs for source/destination refs, digests, and verification state for downstream
   release jobs.

## Core user scenarios

### Priority 1: Full publish orchestration for callers

A maintainer calls the action once with publish settings and trust settings; the action publishes to
source, signs/verifies source digest, optionally promotes to Quay, and returns source/destination
outputs.

**Test coverage:** `.github/workflows/ci.yml` publishes `testdata/minimal-layout` to a local registry and asserts `steps.publish.outputs.digest` is non-empty.

### Priority 2: SDK mode (optional binary)

Once a **`gemara`** (or compatible) CLI is available on the runner `PATH` or as a file path, the maintainer sets `publish_mode: sdk`, `gemara_binary`, and `sdk_args`. The Action logs in when required, exports `GEMARA_REGISTRY`, `GEMARA_REPOSITORY`, and `GEMARA_TAG`, runs the binary, then resolves the digest the same way as layout mode.

**Test coverage:** optional stub or real CLI job (see [tasks.md](tasks.md)); until stable, **layout-copy** remains the recommended path.

### Priority 3: Destination trust behavior

Callers can choose trust mode:

- `copy-only`
- `copy-referrers`
- `resign` (default)

and verify destination trust with the same workflow identity constraints.

### Priority 4: Local or HTTP registry (CI)

For `localhost` registries, the caller sets `plain_http: true` so the Action skips HTTPS login and uses ORAS plain-HTTP flags consistent with the implementation.

## Edge cases addressed

- **Missing layout:** Fail before copy if `index.json` is absent.
- **HTTPS without password:** Fail with a clear error (unless `plain_http: true`).
- **Digest resolution:** Use `oras resolve` on source/destination references and fail fast if unavailable.
- **Username default:** When using password auth, default username behavior matches `action.yml` / README (e.g. `GITHUB_ACTOR` when username omitted).

## Functional requirements summary

The Action must:

1. **Pin ORAS** using the `oras_version` input and install the matching official release for the runner OS/architecture.
2. **Login** with `oras login` for HTTPS when `password` is supplied; **skip login** for `plain_http: true` as implemented.
3. **`layout-copy`:** Require `layout_ref` and a valid layout directory; run `oras cp --from-oci-layout` to `registry/repository:tag`.
4. **`sdk`:** Require `gemara_binary`; export `GEMARA_*` env vars; invoke the binary with `sdk_args`.
5. **Optional promotion:** copy source -> Quay with selected trust mode and destination checks.
6. **Output contract:** append source/destination refs, digests, and verification booleans to
   `GITHUB_OUTPUT`.

## Scope boundaries

**In scope:** ORAS install pin, registry auth glue, publish modes, sign/verify orchestration,
optional promotion, structured outputs.

**Out of scope:** Gemara YAML schema ownership, layer `mediaType` tables, Pack/Unpack
implementation details, vendoring a full SDK publish implementation in bash.

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

### Requirement: Source and destination output contract

After a successful publish, the Action SHALL write source digest/reference outputs. If promotion is
enabled, it SHALL write destination digest/reference outputs and trust/verification statuses.

#### Scenario: Outputs available to downstream steps

- **WHEN** publish completes successfully
- **THEN** `digest`, `source_digest`, and `source_ref` SHALL be non-empty outputs

#### Scenario: Promotion outputs emitted

- **WHEN** `promote_to_destination` is enabled and succeeds
- **THEN** `destination_ref` and `destination_digest` SHALL be emitted and non-empty

### Requirement: Destination trust mode

When promotion is enabled, the Action SHALL honor `trust_mode` (`copy-only`,
`copy-referrers`, `resign`) and SHALL fail for unsupported values.

#### Scenario: Resign destination

- **WHEN** `trust_mode` is `resign`
- **THEN** the destination digest SHALL be signed and verifiable per configured identity policy

## Success metrics

- **CI green:** `layout-copy` integration publishes a fixture and yields a non-empty `digest`.
- **Consumers** can pin `oras_version` and rely on stable `with:` / `outputs.digest` semantics documented in [README.md](../../README.md).
