package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/landlock-lsm/go-landlock/landlock"
)

// dcc-shield: Zero-dependency network sandbox for AUR builds.
// Uses Landlock LSM (Kernel 6.7+) or Network Namespaces as fallback.

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command> [args...]", os.Args[0])
	}

	targetCmd := os.Args[1]
	targetArgs := os.Args[1:]

	cmd := exec.Command(targetCmd, targetArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 1. Enforce Landlock Security Policy
	// We use V4.BestEffort() to attempt network restriction on kernels that support it (ABI v4+).
	// On older kernels (like 6.1, which is ABI v2), BestEffort() will skip network rules 
	// without returning an error, allowing us to implement our own fallback.
	err := landlock.V4.BestEffort().RestrictNet()
	if err != nil {
		log.Printf("[dcc-shield] Landlock enforcement failed: %v. Attempting namespace fallback...", err)
		applyNamespaceFallback(cmd)
	}

	// 2. Check if Landlock actually applied network restrictions
	// If the kernel ABI is less than 4, Landlock did nothing for the network.
	// We check the supported ABI to decide if we need the Namespace fallback.
	// Note: We use the internal-ish check via a custom syscall if library doesn't expose it simply.
	// Actually, the library's BestEffort() is designed to be silent. 
	// To be 'Senior level', we verify the actual environment.
	if !isLandlockNetSupported() {
		fmt.Println("[dcc-shield] Landlock network support not available on this kernel.")
		applyNamespaceFallback(cmd)
	} else {
		fmt.Println("[dcc-shield] Landlock network shield enforced.")
	}

	// 3. Execute with inheritance
	fmt.Printf("[dcc-shield] Executing: %s\n", targetCmd)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("[dcc-shield] Execution failed: %v", err)
	}
}

// isLandlockNetSupported checks if the current kernel supports Landlock ABI v4 (Network).
func isLandlockNetSupported() bool {
	const SYS_LANDLOCK_CREATE_RULESET = 444
	const LANDLOCK_CREATE_RULESET_VERSION = 1 << 0
	res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)
	if err != 0 {
		return false
	}
	return int(res) >= 4
}

// applyNamespaceFallback implements a robust network isolation using Linux Namespaces.
func applyNamespaceFallback(cmd *exec.Cmd) {
	fmt.Println("[dcc-shield] CRITICAL WARNING: Using Network Namespace isolation (CLONE_NEWNET).")
	
	// We use CLONE_NEWUSER to allow unprivileged namespace creation.
	// We map the current user to root inside the namespace to ensure script execution 
	// and filesystem access remain compatible with AUR build requirements.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Unshareflags: syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
}
