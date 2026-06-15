package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// 1. Define the whitelist of allowed environment variables for AUR builds
	whitelist := map[string]bool{
		"PATH":      true,
		"LANG":      true,
		"TERM":      true,
		"MAKEFLAGS": true,
		"PWD":       true,
		"HOME":      true, // Needed for many build scripts, though the directory itself will be restricted
	}

	fmt.Println("[Secrets Layer] Initiating Environment Scrubbing...")

	// 2. Prepare the clean environment
	var cleanEnv []string
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		if whitelist[key] {
			cleanEnv = append(cleanEnv, env)
		}
	}

	// 3. Simulation: Check for a sensitive variable
	// We'll pass a fiktiv secret to the host environment in the shell command
	secretKey := "DCC_SECRET_TOKEN"
	
	fmt.Printf("[Test] Checking for %s in host environment...\n", secretKey)
	if val, exists := os.LookupEnv(secretKey); exists {
		fmt.Printf("[Test] Host environment: %s is VISIBLE (value: %s)\n", secretKey, val)
	} else {
		fmt.Println("[Test] Host environment: Variable not set.")
	}

	fmt.Println("[Test] Swapping to Scrubbed Environment...")

	// In a real dcc-shield, we would use exec.Cmd.Env = cleanEnv
	// For this test, we simulate the perspective of the child process
	foundInClean := false
	for _, env := range cleanEnv {
		if strings.HasPrefix(env, secretKey+"=") {
			foundInClean = true
			break
		}
	}

	if foundInClean {
		fmt.Printf("[Result] %s: LEAKED into sandbox!\n", secretKey)
	} else {
		fmt.Printf("[Result] %s: HIDDEN from sandbox. (SUCCESS)\n", secretKey)
	}

	fmt.Println("[Secrets Layer] Environment audit complete.")
}
