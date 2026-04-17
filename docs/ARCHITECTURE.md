# Architecture: SDK first, Action as glue

This repository is **not** the source of truth for Gemara OCI semantics. That belongs in **[go-gemara](https://github.com/gemaraproj/go-gemara)** and **[go-gemara#60](https://github.com/gemaraproj/go-gemara/issues/60)** (manifest shape, media types, Pack/Unpack, `oras.Copy` from the packed store to a registry).

## Intended split (upstream: go-gemara)

Conceptually:

1. **`bundle.Pack`** (SDK) produces an OCI store / artifact.
2. **`oras.Copy`** (SDK, via `oras-go/v2`) moves that artifact to a **remote registry** with a tag.
3. **`bundle.Unpack`** (SDK) on the consumer pulls and reads it.

The GitHub Action repo should only add what **CI is good at** (implemented as a single composite step in **`action.yml`**, no separate shell scripts):

- Secrets and **`oras login`** (or equivalent) for the registry.
- Pinning **released** tooling (ORAS CLI today; **`gemara` CLI** when published).
- **`GITHUB_OUTPUT`** (`digest`, etc.) for downstream jobs.
- Optional promotion/signing stays in **caller workflows** (e.g. org-infra, complytime-policies).

## Modes in this Action

| Mode | Role | When to use |
|------|------|-------------|
| **`layout-copy`** (default) | Shell **`oras cp --from-oci-layout`** → registry | **Interim**: Pack (or another step) **exports** an **OCI image layout** on disk; this Action only transports it. Same on-disk shape **complyctl** uses for cache ([001 `research.md`](https://github.com/complytime/complyctl/blob/main/specs/001-gemara-native-workflow/research.md) — `oras.Copy` remote→layout; here layout→remote). |
| **`sdk`** | Invoke a **`gemara`** (or other) **binary** from a prior install step | **Target path**: once the SDK ships a stable **`gemara … publish`** (name TBD), callers **`go install`** or download a release and pass **`gemara_binary`** + **`sdk_args`**. The Action does **not** re-encode ORAS flags; it runs the CLI you point at. |

## Why `layout-copy` exists before the SDK CLI is final

It avoids baking **custom `oras push` layer assembly** into this repo (the failure mode we wanted to avoid). It only copies an **already valid** OCI layout. When **`gemara publish`** (or similar) is available, prefer **`publish_mode: sdk`** so **one implementation** owns Pack + Copy.

## Compliance

- **Do not** add Gemara YAML validation or layer `mediaType` tables here — use **go-gemara**.
- **Do** pin versions (`oras_version`, future `gemara` release tags) and document them in release notes.
