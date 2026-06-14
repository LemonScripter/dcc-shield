# dcc-shield: Zero-Dependency AUR Sandbox

`dcc-shield` is a lightweight security wrapper designed to protect Arch Linux users from supply-chain attacks during AUR package builds. It enforces a **"Default-Deny"** network policy on any process it wraps, ensuring that malicious `PKGBUILD` scripts cannot exfiltrate sensitive data (like `~/.ssh` or environment variables).

## Security Architecture

The tool utilizes a **Dual-Layer Causal Enforcement** logic:

1.  **Primary Layer: Landlock LSM (Kernel 6.7+)**
    - Leverages the official `go-landlock` library to enforce network restrictions at the kernel level.
    - Specifically handles `LANDLOCK_ACCESS_NET_CONNECT_TCP` and `LANDLOCK_ACCESS_NET_BIND_TCP` with no allowed rules, creating a total network blackout for the process.
    - Security context is automatically inherited by all child processes (make, gcc, scripts).

2.  **Secondary Layer: Linux Namespaces (Kernel 5.13+)**
    - If Landlock network support is unavailable, the tool transparently falls back to **Network Namespaces** (`CLONE_NEWNET`).
    - The process is executed in a detached network namespace with no interfaces (no `eth0`, no `lo`), making network communication physically impossible.
    - Uses **User Namespaces** (`CLONE_NEWUSER`) with proper UID/GID mapping to ensure full compatibility with unprivileged AUR builds.

## Professional Context

`dcc-shield` implements the **Digital Causal Closure (DCC)** principle. By restricting the network capability at the moment of process creation, we break the causal lánc (chain) required for a data exfiltration attack to succeed. Even if a zero-day exploit allows code execution within the build script, the attacker is trapped in a network-silent environment.

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

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio
