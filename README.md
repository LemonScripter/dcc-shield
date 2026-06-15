# dcc-shield (v2.0): AUR Workflow Causal Enforcer

`dcc-shield` targets the Arch Linux AUR package installation workflow, transforming it into a formally constrained **Digital Causal Closure (DCC)** scope. It reduces the risk of supply-chain exfiltration and malicious build-script behavior by bounding the build/install process within a policy-compliant causal universe.

## Threat Model & Scope (AUR-Specific)

- **Attacker Capabilities:** We assume a malicious AUR `PKGBUILD`, downloaded source, or build-time script executing with the privileges of the build user.
- **In Scope:** `dcc-shield` targets outbound network exfiltration, filesystem tampering, environment variable theft (secrets), SSH key theft, and child-process escape attempts during the package build phase.
- **Boundaries:** The tool is designed to secure the AUR package installation workflow. Host integrity, kernel compromise, firmware exploits, or physical attacks are outside the DCC scope of this installation sandbox. It is not a substitute for full host hardening or manual `PKGBUILD` reviews.

## Multi-Layer DCC Enforcement

The v2.0 engine implements a multi-layer isolation model to ensure the AUR install process remains causally closed. Within the DCC-defined universe, only policy-compliant actions may execute.

1.  **Filesystem Layer (Landlock ABI v1-v2):**
    - Enforces a strict whitelist: Read-Only access to the standard toolchain (`/usr`, `/lib`, `/bin`, `/etc`) and Read/Write access only to the build directory and `/tmp`.
    - Sensitive paths (e.g., `~/.ssh`, `~/.gnupg`) are removed from the process's causal reach.

2.  **Network Layer (Landlock v4 or Namespace Fallback):**
    - Restricts TCP connect/bind capabilities through port-based controls.
    - If Landlock network support is unavailable, the tool uses an isolated network namespace (`CLONE_NEWNET`), detaching the process from the host network.

3.  **Secrets Layer (Environment Scrubbing):**
    - Initiates "Environment Scrubbing" to remove non-essential host variables.
    - Only policy-compliant variables (e.g., `PATH`, `LANG`, `MAKEFLAGS`) are exposed to the build environment.

4.  **Process Layer (Causal Inheritance):**
    - Ensures all child processes (gcc, make, sh) inherit the enforced DCC context.
    - Uses `CLONE_NEWUSER` with UID/GID mapping for compatibility with unprivileged builds.

## Fail-Closed Logic

Security is maintained through strict causal boundaries. If any layer of the DCC framework fails to initialize or encounters a kernel-level error (e.g., unsupported ABI or permission limits), the tool **exits closed** and aborts the installation process to prevent unshielded execution.

## Coverage Matrix

| Attack Vector | Mitigation Layer | Expected Behavior | Test Evidence |
| :--- | :--- | :--- | :--- |
| **Exfiltration via connect()** | Landlock or Namespace | Connection denied / Network unreachable | `strace` confirms `connect()` error |
| **SSH Key / Secret Theft** | Landlock (FS Layer) | Access denied to `~/.ssh` | Verified via `test-sandbox.sh` |
| **Environment Variable Theft** | Secrets Layer | Hidden from sandbox environment | Scrubbing audit successful |
| **Child-process escape** | Process Layer | Restrictions persist in all descendants | Inheritance test pass (Tokyo Server) |
| **Non-whitelisted file changes**| Landlock (FS Layer) | Access denied to non-build paths | Fails by design (Out of Scope) |
| **Kernel/Host compromise** | None (Out of Scope) | Modifications allowed | N/A (Outside DCC scope) |

## Usage

```bash
# Build the static binary
make

# Wrap your AUR helper (creates the DCC universe)
./dcc-shield paru -S target-package
```

## Auditability & Verification

The included test suite provides empirical evidence of the DCC isolation layers.

### What is being verified?
1.  **Syscall Denied:** Uses `strace` to confirm that unauthorized `connect()` or `open()` calls result in a network error or are denied by the kernel.
2.  **Inheritance Proof:** Spawns sub-shells to ensure child processes cannot escape the DCC boundaries.
3.  **Capability Selection:** Verifies that the tool correctly detects kernel capabilities (Landlock ABI version) and selects the appropriate enforcement layer.

## Future Generalization (`dcc-shield-all`)

This repository serves as the AUR-specific foundation for a future generalized `dcc-shield-all` project. While the current scope is formally bounded to the AUR workflow, the broader version will later apply these DCC principles to arbitrary package management systems.

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio
