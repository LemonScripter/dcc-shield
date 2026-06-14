# dcc-shield: Zero-Dependency AUR Sandbox

`dcc-shield` is a lightweight, zero-dependency security wrapper designed to protect your system from supply-chain attacks during AUR package builds. It utilizes the **Linux Landlock LSM** and **Network Namespaces** to enforce a "default-deny" network policy on any process it wraps.

## The Problem: Unvetted PKGBUILDs

AUR (Arch User Repository) helpers like `paru` or `yay` rely on `PKGBUILD` scripts. As highlighted in recent security discussions (e.g., [HUP.hu #190052](https://hup.hu/node/190052)), these scripts can be modified to execute malicious code during the build process. A common attack vector is data exfiltration—stealing your `~/.ssh` keys, environment variables, or private data and sending it to a remote server.

While manual code review is the first line of defense, human error or complex obfuscation can lead to missed threats.

## The Solution: dcc-shield

`dcc-shield` provides a **Zero-Trust safety net**. It ensures that even if a malicious script executes, it has **no way to reach the internet**.

1.  **Zero Dependencies:** A single static binary. No eBPF, no Python, no extra libraries. It runs where your AUR helper runs.
2.  **Kernel Native:** It uses Landlock (Kernel 6.7+) or falls back to Network Namespaces (Kernel 5.13+).
3.  **Default-Deny:** The wrapped process and all its children have **zero** network access.

## Community Context

For the Linux systems engineering community, security should be simple, auditable, and dependency-free. `dcc-shield` implements the **Digital Causal Closure (DCC)** principle: by breaking the network capability before execution, we break the causal chain required for a successful data exfiltration attack.

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

