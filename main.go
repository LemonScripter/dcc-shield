package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsyscall "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

// dcc-shield v2.0: Multi-layer DCC Causal Enforcer for AUR builds.

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

	// --- DCC Layer 1: Environment Scrubbing (Always Active) ---
	scrubEnvironment(cmd)

	// --- DCC Layer 2 & 3: Kernel Capabilities Check ---
	// To prevent kernel security conflicts between Landlock and User Namespaces,
	// we enforce a strict branching logic based on Kernel capabilities.
	if isLandlockNetSupported() {
		// Kernel 6.7+ (Advanced DCC Mode)
		// Full Landlock integration: Filesystem + Network
		
		fmt.Println("[dcc-shield] Kernel 6.7+ detected. Activating Advanced DCC Mode.")
		
		// 1. Enforce Network Blackout
		err := landlock.V4.BestEffort().RestrictNet()
		if err != nil {
			log.Fatalf("[dcc-shield] Network DCC enforcement failed: %v. EXITING CLOSED.", err)
		}

		// 2. Enforce Filesystem Bounding
		rxAccess := landlock.AccessFSSet(llsyscall.AccessFSExecute | llsyscall.AccessFSReadFile | llsyscall.AccessFSReadDir)
		err = landlock.V4.BestEffort().RestrictPaths(
			landlock.PathAccess(rxAccess, "/usr", "/lib", "/lib64", "/bin", "/etc"),
			landlock.RWDirs("/tmp", "."),
		)
		if err != nil {
			log.Fatalf("[dcc-shield] Filesystem DCC enforcement failed: %v. EXITING CLOSED.", err)
		}
		
		fmt.Println("[dcc-shield] Landlock FS + NET bounding enforced.")

	} else {
		// Kernel < 6.7 (Legacy Fallback Mode)
		// We use Network Namespaces. We MUST NOT apply Landlock FS rules here,
		// because an already Landlocked process cannot safely unshare user namespaces
		// (Kernel prevents this to block privilege escalation tricks).
		
		fmt.Println("[dcc-shield] Landlock network support unavailable. Activating Fallback Mode.")
		applyNamespaceFallback(cmd)
	}

	// --- Execution Phase ---
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	fmt.Printf("[dcc-shield] Executing in DCC Universe: %s\n", targetCmd)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("[dcc-shield] Execution failed: %v", err)
	}
}

// scrubEnvironment removes non-essential host variables to prevent secret leakage.
func scrubEnvironment(cmd *exec.Cmd) {
	whitelist := map[string]bool{
		"PATH":      true,
		"LANG":      true,
		"TERM":      true,
		"MAKEFLAGS": true,
		"PWD":       true,
		"HOME":      true,
		"USER":      true,
		"SHELL":     true,
	}

	var cleanEnv []string
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if whitelist[pair[0]] {
			cleanEnv = append(cleanEnv, env)
		}
	}
	cmd.Env = cleanEnv
	fmt.Println("[dcc-shield] Environment scrubbing complete.")
}

// isLandlockNetSupported verifies if Landlock ABI v4 (TCP restrictions) is active.
func isLandlockNetSupported() bool {
	const SYS_LANDLOCK_CREATE_RULESET = 444
	const LANDLOCK_CREATE_RULESET_VERSION = 1 << 0
	res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)
	if err != 0 {
		return false
	}
	return int(res) >= 4
}

// applyNamespaceFallback implements robust network isolation using Linux Namespaces.
func applyNamespaceFallback(cmd *exec.Cmd) {
	fmt.Println("[dcc-shield] Network Namespace isolation enforced (CLONE_NEWNET).")
	
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	
	cmd.SysProcAttr.Unshareflags |= syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER
	cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: os.Getuid(), Size: 1},
	}
	cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: os.Getgid(), Size: 1},
	}
}
