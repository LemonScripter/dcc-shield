# dcc-shield: Zero-Dependency AUR Sandbox

`dcc-shield` is a lightweight, zero-dependency security wrapper designed to protect your system from supply-chain attacks during AUR package builds. It utilizes the **Linux Landlock LSM** (Landlock Security Module) to enforce a "default-deny" network policy on any process it wraps.

## Why dcc-shield?

Supply-chain attacks in AUR (Arch User Repository) helpers like `paru` or `yay` often involve malicious `PKGBUILD` scripts that attempt to exfiltrate data (like `~/.ssh` or environment variables) to a remote server during the build process.

While tools like `Tetragon` or `Falco` are powerful, they require eBPF, kernel headers, or background daemons. `dcc-shield` is different:

1.  **Zero Dependencies:** It's a single static binary. No eBPF, no Python, no extra libraries.
2.  **Kernel Native:** It uses Landlock, which is built into modern Linux kernels (6.7+ for network support).
3.  **Default-Deny:** Unless explicitly allowed (which this version doesn't even implement), the wrapped process has **zero** network access. It can't even `ping` or `curl`.

## Community Context

For the Linux systems engineering community, security should be simple, auditable, and dependency-free. `dcc-shield` implements the **Digital Causal Closure (DCC)** principle at the process level. By restricting network capabilities before the execution starts, we break the causal chain of data exfiltration attacks.

## Requirements

- **Linux Kernel 6.7+** for native Landlock network support.
- **Linux Kernel 5.13+** for fallback mode (Network Namespaces + User Namespaces).
- **Landlock enabled** in your kernel (`lsm=landlock` in boot parameters).

## How to use

1.  **Build the tool:**
    ```bash
    make
    ```
2.  **Wrap your command:**
    ```bash
    ./dcc-shield paru -S some-package
    ```

If the process attempts a network connection on a supported kernel, Landlock will block the syscall. On older kernels, the process will execute in a detached network namespace with no connectivity.

## Technical Details

- **Language:** Go (Static binary, zero dependencies).
- **Primary Mechanism:** `landlock_create_ruleset`, `landlock_restrict_self`.
- **Fallback Mechanism:** `CLONE_NEWNET` | `CLONE_NEWUSER` (Namespace Isolation).

---
**MetaSpace.Bio Logic Engine Project**  
[metaspace.bio](https://metaspace.bio) | admin@metaspace.bio

