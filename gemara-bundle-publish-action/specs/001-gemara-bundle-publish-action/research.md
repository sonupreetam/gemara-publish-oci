# Research: Gemara Bundle Publish Action

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

## 1. SDK pin until PR #62 merges

**Decision**: Keep `replace` in `tools/publish/go.mod` pointing at the **fork pseudo-version** that contains `bundle` APIs until **gemaraproj/go-gemara** ships a tagged release with the same surface.

**Rationale**: Issue [#63](https://github.com/gemaraproj/go-gemara/issues/63) and maintainers expect E2E validation against [PR #62](https://github.com/gemaraproj/go-gemara/pull/62) before merge; consumers need a reproducible pin.

**Alternatives considered**: Vendor a `go.mod` `replace` to a local `go-gemara` clone (harder in CI); depend only on `main` without replace (unstable).

## 2. Digest after publish

**Decision**: After successful `oras.Copy`, call **`remote.Repository.Resolve(ctx, tag)`** and emit **`algorithm:hex`** to `GITHUB_OUTPUT` and stdout (`digest=…`).

**Rationale**: Matches spec clarification; same client as copy; no ORAS CLI parsing.

**Alternatives considered**: Parse `oras cp` stdout (not used—no ORAS CLI); retry loop for eventual consistency (explicitly **out of scope for v1** per spec).

## 3. HTTPS-only v1 composite

**Decision**: No first-class **`plain_http`** input on this composite action for v1.

**Rationale**: Clarification **Option A**; reduces support matrix; [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci) documents `plain_http` for transport-only local CI if needed.

**Alternatives considered**: Mirror `plain_http` from gemara-publish-oci in this repo (rejected for v1 scope).

## 4. Single-step vs two-phase publish

**Decision**: Primary path remains **root YAML → memory pack → `oras.Copy`**. Document **layout + `oras cp`** as a **separate** flow using [gemara-publish-oci](https://github.com/sonupreetam/gemara-publish-oci) `layout-copy` (or future SDK disk export) without duplicating pack logic in transport.

**Rationale**: Constitution **V** and User Story 4.

**Alternatives considered**: Add a second `publish_mode` to this action (deferred—keeps action.yml smaller; transport repo already exists).

## 5. Registry compatibility caveats

**Decision**: Document in README/plan that registries or clients expecting **only** legacy single-layer Gemara media types may not accept **bundle** manifests until aligned ([spec Open Questions](./spec.md#open-questions)).

**Rationale**: No action-side workaround without violating SDK-owned contract.

**Alternatives considered**: Transcode bundle to legacy layout in the action (rejected—violates FR-002).
