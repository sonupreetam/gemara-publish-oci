# Plan: 002 — Gemara OCI supply chain (handoff, no 001 spec edits)

**Spec:** [spec.md](./spec.md)

## Objective

Keep **001** the **action** product spec; use **002** for **where this repo sits** in
**complytime-policies** / **org-infra** / **Quay** E2E without amending [001
spec.md](../001-gemara-bundle-publish-action/spec.md).

## Phases

1. **Align links** when **complytime/oci-artifact** hosts this **002** path on `main`.
2. **No change** to composite **`action.yml`** for Quay; confirm with **org-infra** **008** when
   promote design is final.
3. **Close 002** when the team agrees the [integration.md](./integration.md) boundary is the recorded
   understanding.

## Cross-repo

- [complytime-policies
  002](https://github.com/complytime/complytime-policies/blob/main/specs/002-policy-oci-quay-e2e-supply-chain/spec.md)
- [complytime/org-infra
  008](https://github.com/complytime/org-infra/blob/main/specs/008-quay-promote-signature-verification/spec.md)
