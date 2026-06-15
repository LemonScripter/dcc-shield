package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		runChild()
		return
	}

	fmt.Println("[Process Layer] Testing PID Namespaces (No Landlock)...")

	executable, _ := os.Executable()

	// 1. Spawn child with PID & User Namespace
	cmd := exec.Command(executable, "child")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Unshareflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("[Parent] Spawning sandboxed child with NEWPID...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("[Parent] Child failed: %v\n", err)
	}
}

func runChild() {
	fmt.Printf("[Child] My PID is %d\n", os.Getpid())
	
	if os.Getpid() == 1 {
		fmt.Println("[Child] PID Namespace isolation: SUCCESS (I am PID 1).")
	} else {
		fmt.Printf("[Child] PID Namespace isolation: FAILED (PID is %d).\n", os.Getpid())
	}
}
