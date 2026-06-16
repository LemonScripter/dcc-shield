# dcc-shield (v2.0): AUR Workflow Causal Enforcer

[![Verified](https://img.shields.io/badge/Verified-Tokyo--Node-green)](https://github.com/LemonScripter/dcc-shield/blob/master/VERIFICATION.md)
[![Status](https://img.shields.io/badge/Status-Hardened--STABLE-blue)](README.md)
[![Project](https://img.shields.io/badge/BioOS-Causal--Security-green)](https://bioos.metaspace.bio)
[![DOI](https://img.shields.io/badge/DOI-10.5281%2Fzenodo.20384700-purple)](https://doi.org/10.5281/zenodo.20384700)
[![Media](https://img.shields.io/badge/BioOS-Causal--Security-orange)](https://open.substack.com/pub/szokelaci/p/when-trust-becomes-an-attack-surface)

`dcc-shield` targets the Arch Linux AUR package installation workflow and constrains it within a Digital Causal Closure (DCC)-defined scope. It reduces the risk of supply-chain exfiltration and malicious build-script behavior during AUR installs.

## Scientific Foundation and Lineage

The design and enforcement mechanics of this tool are direct practical implementations of the author's broader research into Digital Causal Closure (DCC). These background materials serve as the conceptual foundation and architectural lineage for the tool:

- **Research Paper:** The Causal Operating System: Digital Causal Closure for Autonomous Systems ([DOI: 10.5281/zenodo.20384700](https://doi.org/10.5281/zenodo.20384700))
- **Formal Specification:** [BioOS Causal Constitution (PDF)](https://bioos.metaspace.bio/bioos_causal_constitution_en.pdf)

## Scope and Threat Model

This tool's current scope is AUR package installation on Arch Linux. It is designed to mitigate risks during the build/install phase, such as malicious `PKGBUILD` scripts or compromised downloaded source code. Manual `PKGBUILD` review remains valuable; this tool is not a substitute for it.

**In-Scope Threats:**
- Outbound network exfiltration
- Filesystem tampering
- SSH key and secret theft
- Environment variable theft
- Child-process escape attempts

**Out-of-Scope Threats:**
- Compromised kernel or firmware attacks
- Physical intrusion
- General host hardening
- Trusting the system on an already compromised machine

*(See [THREAT_MODEL.md](THREAT_MODEL.md) for further details.)*

## Hardened Architecture

The enforcement relies on a multi-layer isolation model:

- **Filesystem Layer:** Mediated by the available Landlock ABI and a project-defined allowlist.
- **Network Layer:** Enforced by native Landlock network capabilities where available; otherwise, it uses an isolated network namespaces fallback.
- **Secrets Layer:** Utilizes explicit allowlist-based environment scrubbing.
- **Process Layer:** Ensures descendant processes inherit the same enforcement context. It uses `CLONE_NEWUSER` with UID/GID mapping as a compatibility mechanism for unprivileged builds in namespace fallback mode.

## Smart Fallback and Lifecycle

The tool detects kernel capabilities and chooses the enforcement path accordingly. If native enforcement is unavailable or a kernel-level error occurs, the tool exits closed and aborts the installation.

`dcc-shield` operates as a transient wrapper. It runs only for the duration of the build/install process and leaves no background daemon or permanent service running.

## Coverage Matrix

| Attack Vector | Mitigation Layer | Expected Behavior | Test Evidence |
| :--- | :--- | :--- | :--- |
| **Exfiltration via connect()** | Landlock or Namespace | Connection denied / Network unreachable | Empirical trace confirms `connect()` error |
| **SSH Key / Secret Theft** | Landlock (FS Layer) | Access denied to `~/.ssh` | Verified behavior under test via `test-sandbox.sh` |
| **Environment Variable Theft** | Secrets Layer | Hidden from sandbox environment | Scrubbing audit confirms behavior |
| **Child-process escape** | Process Layer | Restrictions persist in descendant processes | Verified behavior under test via sub-shell testing |

## Enforcing Causal Boundaries (PAE Concept)

The design focuses on enforcing causal boundaries rather than maintaining package signatures, reducing dependence on per-package blacklists. 

**Positive Axiom Exclusion (PAE)** is a conceptual framing for this enforcement model, meaning only explicitly permitted causal paths are allowed. During the June 2026 "Atomic Arch" supply chain attack, malicious payloads were delivered via orphaned AUR packages. By restricting network and filesystem access during the build phase, this enforcement model mitigates such payload delivery mechanisms without requiring per-package signature updates.

## Forensic Report

A detailed forensic report of the test setup and empirical validation is available in the [Atomic Arch Forensic Report](proofs/atomic-arch/REPORT.md). This report provides reproducible evidence of the observed behavior under test conditions.

## Auditability and Verification

The included test suite provides empirical evidence of the isolation layers. It verifies that:
1. Unauthorized syscalls are denied or fail under isolation.
2. Child processes inherit the DCC boundaries.
3. Kernel capabilities are detected and the correct enforcement path is selected.

## Project Positioning

This is an open-source project currently focused on AUR-specific workflows. The long-term direction is to generalize this enforcement model to broader package-installation workflows once the AUR implementation is stabilized.

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio
