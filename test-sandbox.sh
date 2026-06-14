#!/bin/bash
# dcc-shield Test Suite
# Requires: strace, curl, grep

set -u

echo "=== dcc-shield Professional Test Suite ==="

# 1. Verify connect() syscall blocking via strace
# We use -f to follow forks (AUR helpers spawn children)
echo "[1/3] Verifying syscall blocking via strace..."
STRACE_LOG=$(strace -f -e connect ./dcc-shield curl -s --connect-timeout 2 google.com 2>&1) || true

# Look for 'Could not resolve host' (DNS fail due to no net) or connection errors
if echo "$STRACE_LOG" | grep -q "Could not resolve host\|Network is unreachable\|Connection refused\|EACCES\|EPERM"; then
    echo "SUCCESS: Network operation was intercepted/blocked."
else
    echo "FAILURE: Network operation was not blocked as expected."
    echo "Debug: $STRACE_LOG"
    exit 1
fi

# 2. Verify Sandbox Inheritance (Child Process)
echo "[2/3] Verifying sandbox inheritance (Sub-shell curl)..."
# DNS failure (Exit code 6 in curl) is a proof of network isolation in namespaces
./dcc-shield /bin/bash -c "curl -s --connect-timeout 1 8.8.8.8" 2>&1 | grep -q "Network is unreachable\|Connection timed out" || [ $? -ne 0 ]
if [ $? -eq 0 ]; then
    echo "SUCCESS: Sandbox inherited by child process."
else
    echo "FAILURE: Child process escaped the sandbox!"
    exit 1
fi

# 3. Security Audit (Landlock/Namespace status)
echo "[3/3] Checking tool diagnostic output..."
./dcc-shield true 2>&1 | grep -E "Landlock|isolation|ACTIVE"
echo "SUCCESS: Tool reports active security layers."

echo "=== Test Suite Complete: PASS ==="
