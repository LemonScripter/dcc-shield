# dcc-shield: Zero-Dependency AUR Sandbox

`dcc-shield` is a lightweight security wrapper designed to reduce the risk of supply-chain exfiltration attacks during AUR package builds. It enforces a "Default-Deny" network policy on wrapped processes, limiting the ability of `PKGBUILD` scripts to exfiltrate sensitive data (such as `~/.ssh` or environment variables) over the network.

## Threat Model & Scope

- **Attacker Capabilities:** We assume the attacker has successfully injected malicious code into an AUR `PKGBUILD` or its downloaded source, which executes under the build user’s privileges.
- **In Scope:** `dcc-shield` specifically targets **outbound network exfiltration** during the execution of the wrapped command. 
- **Out of Scope:** `dcc-shield` is **not** a replacement for manual `PKGBUILD` reviews. It does **not** restrict filesystem modifications. An attacker can still tamper with the built package, drop persistent backdoors, or alter local files accessible to the build user. Filesystem tampering, local persistence, and built-package modifications are explicitly out of scope.

## Security Architecture & Failure Modes

The tool utilizes a dual-layer network enforcement design. It is engineered to fail closed on critical setup errors but falls back gracefully across isolation layers where appropriate.

1.  **Primary Layer: Landlock LSM (Kernel 6.7+)**
    - Leverages the `go-landlock` library to restrict TCP capabilities.
    - Specifically restricts `LANDLOCK_ACCESS_NET_CONNECT_TCP` and `LANDLOCK_ACCESS_NET_BIND_TCP` through port-based controls.
    - **Note:** This is a capability-based restriction, not domain/IP allowlisting.
    - The enforced security context is inherited by all child processes.
    - **Failure Mode:** If Landlock setup fails or is unsupported by the kernel, the tool attempts the namespace fallback.

2.  **Secondary Layer: Linux Namespaces (Kernel 5.13+)**
    - If Landlock network support is unavailable, the tool uses an isolated network namespace (`CLONE_NEWNET`).
    - The wrapped process is not exposed to external network interfaces and is isolated from the host network.
    - Uses `CLONE_NEWUSER` with UID/GID mapping for compatibility with unprivileged builds.
    - **Failure Mode:** If the namespace fallback also fails, the tool **exits closed** and aborts the build process to prevent unshielded execution.

## Coverage Matrix

| Attack Vector | Mitigation Layer | Expected Behavior | Test Evidence |
| :--- | :--- | :--- | :--- |
| **Exfiltration via connect()** | Landlock or Namespace | Connection denied / Network unreachable | `strace` confirms `connect()` error |
| **Child-process inheritance** | Landlock or Namespace | Restrictions persist in spawned sub-shells | Verified via `test-sandbox.sh` |
| **Landlock unavailable** | Fallback to `CLONE_NEWNET` | Executes in isolated namespace | Kernel capability detection test |
| **Non-network file changes** | None (Out of Scope) | Modifications allowed | Fails by design (Filesystem is out of scope) |

## Usage

```bash
# Build the static binary
make

# Wrap your AUR helper
./dcc-shield paru -S target-package
```

## Auditability & Verification

For a security tool to be credible, its enforcement must be verifiable. `dcc-shield` includes a test suite to provide empirical evidence of isolation.

### Running the Test Suite
The included `test-sandbox.sh` script automates the verification process:

```bash
# Requirements: strace, curl, grep
chmod +x test-sandbox.sh
./test-sandbox.sh
```

### What is being verified?
1.  **Syscall Denied:** Uses `strace` to confirm that the `connect()` syscall is denied or results in a network error.
2.  **Inheritance Proof:** Spawns a sub-shell and attempts a network operation to ensure that child processes cannot escape the sandbox.
3.  **Capability Selection:** Verifies that the tool correctly detects kernel capabilities and selects either Landlock or the namespace fallback as appropriate.

## Feedback & Contributions

This project serves as a practical demonstration of the **Digital Causal Closure (DCC)** principle. The source code is open for audit and the tests are fully automated. We actively welcome contributions and feedback regarding the refinement of Landlock rules or the integration of other kernel-native isolation mechanisms, such as `seccomp-bpf`.

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio
