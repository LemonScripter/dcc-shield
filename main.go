package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// Landlock and Prctl constants
const (
	SYS_LANDLOCK_CREATE_RULESET = 444
	SYS_LANDLOCK_ADD_RULE        = 445
	SYS_LANDLOCK_RESTRICT_SELF   = 446

	LANDLOCK_CREATE_RULESET_VERSION = 1 << 0
	LANDLOCK_RULE_NET_PORT          = 2

	LANDLOCK_ACCESS_NET_BIND_TCP    = 1 << 0
	LANDLOCK_ACCESS_NET_CONNECT_TCP = 1 << 1

	PR_SET_NO_NEW_PRIVS = 38
)

type landlockRulesetAttr struct {
	HandledAccessFs  uint64
	HandledAccessNet uint64
}

type landlockNetPortAttr struct {
	AllowedAccess uint64
	Port          uint64
}

// getLandlockABI queries the supported Landlock ABI version.
func getLandlockABI() int {
	res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)
	if err != 0 {
		return 0
	}
	return int(res)
}

// restrictSelf enforces the created ruleset on the calling process and its future children.
func restrictSelf(rulesetFd uintptr) error {
	// PR_SET_NO_NEW_PRIVS is a prerequisite for Landlock if the process is not privileged.
	_, _, err := syscall.Syscall(syscall.SYS_PRCTL, PR_SET_NO_NEW_PRIVS, 1, 0)
	if err != 0 {
		return fmt.Errorf("failed to set no_new_privs: %v", err)
	}

	_, _, err = syscall.Syscall(SYS_LANDLOCK_RESTRICT_SELF, rulesetFd, 0, 0)
	if err != 0 {
		return fmt.Errorf("failed to enforce ruleset (restrict_self): %v", err)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "dcc-shield: Zero-dependency network sandbox wrapper\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	abi := getLandlockABI()
	useLandlockNet := abi >= 4

	fmt.Printf("[dcc-shield] Initializing security layers (Kernel ABI v%d)...\n", abi)

	// Setup Command
	targetCmd := os.Args[1]
	targetArgs := os.Args[1:]
	cmd := exec.Command(targetCmd, targetArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if useLandlockNet {
		// 1. Create Ruleset
		// We explicitly handle both Bind and Connect to ensure a default-deny policy.
		attr := landlockRulesetAttr{
			HandledAccessNet: LANDLOCK_ACCESS_NET_BIND_TCP | LANDLOCK_ACCESS_NET_CONNECT_TCP,
		}
		
		res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), 16, 0)
		if err != 0 {
			fmt.Fprintf(os.Stderr, "[dcc-shield] CRITICAL: landlock_create_ruleset failed: %v\n", err)
			os.Exit(1)
		}
		rulesetFd := uintptr(res)
		defer syscall.Close(int(rulesetFd))

		// 2. Populate Ruleset (Optional)
		// To block all network, we simply add NO rules to the ruleset.
		// However, to demonstrate 'landlock_add_rule' implementation:
		// Example: Allowing localhost/port 80 would look like this:
		/*
		portAttr := landlockNetPortAttr{
			AllowedAccess: LANDLOCK_ACCESS_NET_CONNECT_TCP,
			Port:          80,
		}
		_, _, err = syscall.Syscall(SYS_LANDLOCK_ADD_RULE, rulesetFd, LANDLOCK_RULE_NET_PORT, uintptr(unsafe.Pointer(&portAttr)))
		*/

		// 3. Restrict Self
		if err := restrictSelf(rulesetFd); err != nil {
			fmt.Fprintf(os.Stderr, "[dcc-shield] CRITICAL: Security enforcement failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[dcc-shield] Landlock network shield ACTIVE.")

	} else {
		// Fallback for older kernels (ABI < 4)
		fmt.Println("[dcc-shield] WARNING: Landlock network support requires Kernel 6.7+. Falling back to Namespaces.")
		
		// Map current UID/GID to Root inside the namespace to maintain filesystem access for AUR builds.
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Unshareflags: syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{
				{ContainerID: 0, HostID: os.Getuid(), Size: 1},
			},
			GidMappings: []syscall.SysProcIDMap{
				{ContainerID: 0, HostID: os.Getgid(), Size: 1},
			},
		}
		fmt.Println("[dcc-shield] CLONE_NEWNET network isolation ACTIVE.")
	}

	fmt.Printf("[dcc-shield] Executing target: %s\n", targetCmd)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "[dcc-shield] Execution failed: %v\n", err)
		os.Exit(1)
	}
}
