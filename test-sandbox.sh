#!/bin/bash
# dcc-shield Test Suite
# Requires: strace, curl

set -e

echo "=== dcc-shield Professional Test Suite ==="

# 1. Verify connect() syscall blocking via strace
echo "[1/3] Verifying syscall blocking via strace..."
strace -e connect ./dcc-shield curl -s --connect-timeout 2 google.com 2>&1 | grep "EACCES\|EPERM\|ECONNREFUSED\|ENETUNREACH" > /dev/null
if [ $? -eq 0 ]; then
    echo "SUCCESS: connect() syscall was intercepted and blocked."
else
    echo "FAILURE: connect() syscall was not blocked as expected."
fi

# 2. Verify Sandbox Inheritance (Child Process)
echo "[2/3] Verifying sandbox inheritance (Sub-shell ping)..."
./dcc-shield /bin/bash -c "ping -c 1 8.8.8.8" 2>&1 | grep "Operation not permitted\|Network is unreachable" > /dev/null
if [ $? -eq 0 ]; then
    echo "SUCCESS: Sandbox inherited by child process."
else
    echo "FAILURE: Child process escaped the sandbox!"
fi

# 3. Check dmesg for Landlock Audit Logs
echo "[3/3] Checking dmesg for Landlock audit events..."
# Note: This requires sudo or specific kernel permissions.
if command -v sudo &> /dev/null; then
    sudo dmesg | grep -i "landlock" | tail -n 5 || echo "INFO: No recent Landlock logs in dmesg (normal if no violations were logged to kmsg)."
else
    echo "INFO: Skipping dmesg check (sudo not available)."
fi

echo "=== Test Suite Complete ==="
