# Feature specification: Gemara OCI action — supply chain and downstream Quay (E2E handoff)

**Feature ID:** `002-gemara-oci-supply-chain-e2e`  
**Created:** 2026-04-23  
**Status:** Draft  
**Relates to:** [001-gemara-bundle-publish-action](../001-gemara-bundle-publish-action/spec.md) (composite
contract and **action.yml** scope **unchanged** by this feature)

## Purpose

Document how **this repository** (composite **oci-artifact** / **gemara-publish-oci** transport)
sits in the **ComplyTime policy OCI** chain (GHCR → sign → **Quay**) and which concerns are **downstream**
in **complytime/org-infra** and **complytime-policies**. This is a **sibling** handoff to
[complytime-policies
002](https://github.com/complytime/complytime-policies/blob/main/specs/002-policy-oci-quay-e2e-supply-chain/spec.md)
and [org-infra
008](https://github.com/complytime/org-infra/blob/main/specs/008-quay-promote-signature-verification/spec.md).

## In scope

- **Boundary:** This action = **root YAML → pack → `oras.Copy` to registry** only.
- **Out of this repo:** SLSA policy, keyless **cosign**, **Quay promote** — org-infra reusables; thin
  caller — complytime-policies.
- **Open:** Quay **dest** `cosign verify` / **`oras copy -r`** — [org-infra
  008](https://github.com/complytime/org-infra/blob/main/specs/008-quay-promote-signature-verification/spec.md).

## Out of scope

- Changing [001](../001-gemara-bundle-publish-action/spec.md) **FR**s or the composite’s HTTPS-only
  and digest contract.
- Embeddng promote logic in `action.yml`.

## References

- [integration.md](./integration.md) — table and team asks.  
- [plan.md](./plan.md) — phases.
