# Threat Model and Scope

`dcc-shield` is an AUR workflow causal enforcer designed to reduce the risk of supply-chain exfiltration and malicious build-script behavior during Arch Linux AUR package installations.

## Target Environment
Arch Linux systems building and installing packages from the Arch User Repository (AUR) via helpers (e.g., `paru`, `yay`) or `makepkg`.

## Attacker Capabilities
The attacker is assumed to have compromised a `PKGBUILD` script, its downloaded source archive, or an installation hook. The attacker can execute arbitrary code during the build/install phase, spawn child processes, and attempt to access the local environment.

## In-Scope Threats
`dcc-shield` applies a policy-compliant causal universe to mitigate the following actions during the build process:
- **Outbound Network Exfiltration:** Attempts to download secondary payloads or upload stolen data.
- **Filesystem Tampering:** Unauthorized writes to system directories outside of `/tmp` or the designated build directory.
- **Credential Theft:** Unauthorized reads of sensitive user directories (e.g., `~/.ssh`, `~/.gnupg`, `~/.aws`).
- **Environment Variable Theft:** Harvesting sensitive tokens or API keys passed in the environment.
- **Child-Process Escape:** Spawning daemonized or orphaned processes to bypass temporal build restrictions.

## Out-of-Scope Threats
This tool is **not** a general host-hardening solution. It provides no guarantees against:
- **Compromised Kernel:** Vulnerabilities in the Linux kernel that allow privilege escalation or namespace/Landlock escapes.
- **Firmware Attacks:** Compromised hardware or boot chains.
- **Physical Intrusion:** Direct access to the machine.
- **General Host Hardening:** Protecting services or user sessions outside of the wrapped AUR build process.
- **Pre-Existing Compromise:** Securing a system that is already infected with a persistent rootkit or malware.

*Note: Manual `PKGBUILD` review remains valuable. `dcc-shield` provides an isolation layer, but it is not a substitute for auditing the code you install.*