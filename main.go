package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// Landlock constants for Linux
const (
	SYS_LANDLOCK_CREATE_RULESET = 444
	SYS_LANDLOCK_RESTRICT_SELF  = 446

	LANDLOCK_CREATE_RULESET_VERSION = 1 << 0

	LANDLOCK_ACCESS_NET_BIND_TCP    = 1 << 0
	LANDLOCK_ACCESS_NET_CONNECT_TCP = 1 << 1
)

type landlockRulesetAttr struct {
	HandledAccessFs  uint64
	HandledAccessNet uint64
}

func getLandlockABI() int {
	res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)
	if err != 0 {
		return 0
	}
	return int(res)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s paru -S google-chrome\n", os.Args[0])
		os.Exit(1)
	}

	abi := getLandlockABI()
	useLandlockNet := abi >= 4

	if !useLandlockNet {
		fmt.Printf("[dcc-shield] Landlock ABI v%d (Network support requires v4+).\n", abi)
		fmt.Println("[dcc-shield] Falling back to Network Namespaces (CLONE_NEWNET) for isolation.")
	} else {
		fmt.Printf("[dcc-shield] Using Landlock ABI v%d for network isolation.\n", abi)
	}

	// 1. Prepare Command
	targetCmd := os.Args[1]
	targetArgs := os.Args[1:]

	cmd := exec.Command(targetCmd, targetArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 2. Apply Network Isolation
	if useLandlockNet {
		// Use Landlock for Network (Kernel 6.7+)
		var attr landlockRulesetAttr
		attr.HandledAccessNet = LANDLOCK_ACCESS_NET_BIND_TCP | LANDLOCK_ACCESS_NET_CONNECT_TCP
		
		res, _, err := syscall.Syscall(SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), 16, 0)
		if err != 0 {
			fmt.Fprintf(os.Stderr, "[dcc-shield] Error creating Landlock ruleset: %v\n", err)
			os.Exit(1)
		}
		fd := uintptr(res)
		defer syscall.Close(int(fd))

		_, _, err = syscall.Syscall(SYS_LANDLOCK_RESTRICT_SELF, fd, 0, 0)
		if err != 0 {
			fmt.Fprintf(os.Stderr, "[dcc-shield] Error enforcing Landlock sandbox: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Use Network Namespace + User Namespace (Reliable fallback for all modern kernels)
		// We map the current user to root inside the namespace to allow executing scripts.
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Unshareflags: syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getuid(),
					Size:        1,
				},
			},
			GidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getgid(),
					Size:        1,
				},
			},
		}
	}

	fmt.Printf("[dcc-shield] Executing '%s'...\n", targetCmd)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "[dcc-shield] Command failed: %v\n", err)
		os.Exit(1)
	}
}
