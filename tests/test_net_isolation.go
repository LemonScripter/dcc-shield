package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		runChild()
		return
	}

	fmt.Println("[Network Layer] Testing Namespace Fallback (CLONE_NEWNET)...")

	executable, _ := os.Executable()

	// 1. Spawn child with Network & User Namespace
	cmd := exec.Command(executable, "child")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Unshareflags: syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("[Parent] Spawning child in isolated network namespace...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("[Parent] Child execution failed: %v\n", err)
	}
}

func runChild() {
	fmt.Println("[Child] Attempting network access (http://google.com)...")
	
	// Try a simple HTTP GET
	client := &http.Client{}
	resp, err := client.Get("http://google.com")
	
	if err != nil {
		fmt.Printf("[Result] Network access: DENIED (%v). SUCCESS.\n", err)
	} else {
		fmt.Printf("[Result] Network access: GRANTED (Status: %s). FAILURE.\n", resp.Status)
		resp.Body.Close()
	}
}
