# Supply chain integration: oci-artifact in the ComplyTime policy OCI chain

**Spec:** [spec.md](./spec.md) | **001 product spec:** [../001-gemara-bundle-publish-action/spec.md](../001-gemara-bundle-publish-action/spec.md)

## Role of this repository

- The composite implements **one** link: **root Gemara YAML → `bundle.Assemble` / `bundle.Pack` →
  `oras.Copy` to a registry** (`action.yml`, `tools/publish/main.go`). It does **not** implement org-wide
  **SLSA**, **SPDX** policy, keyless **cosign**, or **Quay promote**; those are **complytime/org-infra**
  reusables that **embed** this action (or a fork).
- The **OCI manifest** is a **Gemara bundle** per SDK contract, not necessarily a **Docker** image.
  Registry UIs (e.g. Quay) may not show familiar “layers”; **CLI** is ground truth.

## Upstream / downstream

| Component | Responsibility |
|-----------|------------------|
| **go-gemara** | Bundle → OCI contract. |
| **oci-artifact (this repo)** | Pack + push. |
| **complytime/org-infra** | Wrap composite + **GHCR** + attest + sign/verify + **resuable_publish_quay**. |
| **complytime-policies** | Thin caller: `publish-policy-oci.yml` → `workflow_call` to org-infra. |

## Downstream design (not decided here)

- **Quay promote** + **`cosign verify` on destination** — [org-infra
  008](https://github.com/complytime/org-infra/blob/main/specs/008-quay-promote-signature-verification/spec.md)
  (e.g. **cosign copy** vs **`oras copy -r`** for referrer graph on Quay).
- **Cosign graph** vs **ORAS/crane** for out-of-graph artifacts — org policy; this action only supplies
  the **subject** image/bundle to staging as embedded by **reusable_publish_oras**.

## Team (mirrors complytime-policies 002 handoff)

- Review [complytime-policies#5](https://github.com/complytime/complytime-policies/issues/5) and
  [org-infra#211](https://github.com/complytime/org-infra/pull/211) as **architecture** SSOT.
- The **pin** of **oci-artifact** inside `reusable_publish_oras.yml` changes when the org agrees on a
  new `complytime/oci-artifact@…` ref; **complytime-policies** **FR-006** / README then bumps the
  **org-infra** caller SHA.
