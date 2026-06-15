package main

import (
	"fmt"
	"os"

	"github.com/landlock-lsm/go-landlock/landlock"
)

func main() {
	// 1. Define the security policy
	// RODirs access to /usr, /lib, /etc, and /bin
	// RWDirs access to /tmp and current directory (.)
	// EVERYTHING ELSE is blocked (including ~/.ssh)
	
	err := landlock.V2.BestEffort().RestrictPaths(
		landlock.RODirs("/usr", "/lib", "/lib64", "/etc", "/bin"),
		landlock.RWDirs("/tmp", "."),
	)

	if err != nil {
		fmt.Printf("Landlock restriction failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[Sandbox] Landlock filesystem policy enforced (ABI v2).")

	// 2. Test access
	testFile("/etc/passwd") // Should succeed
	
	home := os.Getenv("HOME")
	if home != "" {
		testFile(home + "/.ssh") // Should fail
	} else {
		fmt.Println("[Test] HOME env not set, skipping SSH test.")
	}
}

func testFile(path string) {
	_, err := os.Open(path)
	if err != nil {
		fmt.Printf("[Test] Access to %s: DENIED (%v)\n", path, err)
	} else {
		fmt.Printf("[Test] Access to %s: GRANTED\n", path)
	}
}
