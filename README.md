# dcc-shield: Zero-Dependency AUR Sandbox

`dcc-shield` is a lightweight security wrapper designed to protect Arch Linux users from supply-chain exfiltration attacks during AUR package builds. It enforces a "Default-Deny" network policy on any process it wraps, ensuring that `PKGBUILD` scripts cannot trivially exfiltrate sensitive data (like `~/.ssh` or environment variables) over the network.

## Threat Model & Scope

- **Attacker Capabilities:** We assume the attacker has successfully injected malicious code into an AUR `PKGBUILD` or its downloaded source, which executes with the privileges of the build user.
- **In Scope:** `dcc-shield` specifically targets **network exfiltration** during the execution of the wrapped command. 
- **Out of Scope:** `dcc-shield` is **not** a complete replacement for manual `PKGBUILD` reviews. It does **not** restrict filesystem modifications. An attacker can still tamper with the built package, drop persistent backdoors, or alter local files accessible to the build user. The tool only restricts outbound network connectivity during the build phase.

## Security Architecture & Failure Modes

The tool utilizes a Dual-Layer network enforcement logic. It is designed to fail closed on execution errors but falls back gracefully across isolation layers.

1.  **Primary Layer: Landlock LSM (Kernel 6.7+)**
    - Leverages the `go-landlock` library to restrict TCP network capabilities.
    - Explicitly drops `LANDLOCK_ACCESS_NET_CONNECT_TCP` and `LANDLOCK_ACCESS_NET_BIND_TCP` rights for the process. This restricts port-based TCP connectivity, irrespective of domains or IP addresses.
    - Security context is inherited by all child processes.
    - **Failure Mode:** If Landlock enforcement partially succeeds or if the ABI version is unsupported, the tool logs the limitation and automatically falls back to the Secondary Layer.

2.  **Secondary Layer: Linux Namespaces (Kernel 5.13+)**
    - If Landlock network support is unavailable, the tool falls back to Network Namespaces (`CLONE_NEWNET`).
    - The process is executed in a detached network namespace without external interfaces (no `eth0`), isolating it from the host network.
    - Uses User Namespaces (`CLONE_NEWUSER`) with UID/GID mapping to ensure compatibility with unprivileged builds.
    - **Failure Mode:** If the namespace fallback setup fails (e.g., due to permission limits), the underlying execution will fail, causing the tool to **fail closed** and abort the build process securely.

## Coverage Matrix

| Attack Vector | Mitigation Layer | Expected Behavior | Test Evidence |
| :--- | :--- | :--- | :--- |
| **Exfiltration via connect()** | Landlock or Namespace | Connection refused / Network unreachable | `strace` confirms `connect()` fails |
| **Child-process inheritance** | Landlock or Namespace | Restrictions persist in spawned sub-shells | Verified via `test-sandbox.sh` |
| **Landlock unavailable** | Fallback to `CLONE_NEWNET` | Executes in isolated namespace | Kernel ABI fallback logic tested |
| **Non-network file changes** | None (Out of Scope) | Modifications allowed | N/A |

## Usage

```bash
# Build the static binary
make

# Wrap your AUR helper
./dcc-shield paru -S target-package
```

## Auditability & Verification

For a security tool to be trustworthy, its enforcement must be verifiable. `dcc-shield` includes a professional test suite to provide empirical proof of isolation.

### Running the Test Suite
The included `test-sandbox.sh` script automates the verification process:

```bash
# Requirements: strace, curl, grep
chmod +x test-sandbox.sh
./test-sandbox.sh
```

### What is being verified?
1.  **Syscall Interception:** Uses `strace` to confirm that the `connect()` syscall is physically blocked or results in a network error (e.g., DNS failure due to isolation).
2.  **Inheritance Proof:** Spawns a sub-shell and attempts a network operation to ensure that child processes (like those spawned by `paru` or `make`) cannot escape the sandbox.
3.  **Kernel Integration:** Validates that the tool correctly identifies the kernel's Landlock ABI version and applies the appropriate security layer (Landlock or Namespaces).

## Feedback & Contributions

This project serves as a practical demonstration of the **Digital Causal Closure (DCC)** principle. The source code is open for audit and the tests are fully automated. We actively welcome contributions and feedback regarding the refinement of Landlock rules or the integration of other kernel-native isolation mechanisms, such as `seccomp-bpf`.

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio
