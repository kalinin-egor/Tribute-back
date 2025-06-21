#!/bin/bash

echo "🔍 SSH Configuration Test Script"
echo "=================================="

# Check SSH service status
echo "1. Checking SSH service status..."
systemctl status sshd --no-pager

# Check SSH configuration
echo -e "\n2. Checking SSH configuration..."
echo "PubkeyAuthentication setting:"
grep -i "pubkeyauthentication" /etc/ssh/sshd_config || echo "Not found (defaults to yes)"

echo "PasswordAuthentication setting:"
grep -i "passwordauthentication" /etc/ssh/sshd_config || echo "Not found (defaults to yes)"

echo "PermitRootLogin setting:"
grep -i "permitrootlogin" /etc/ssh/sshd_config || echo "Not found (defaults to yes)"

# Check SSH directory and keys
echo -e "\n3. Checking SSH directory and keys..."
echo "SSH directory permissions:"
ls -la ~/.ssh/ 2>/dev/null || echo "SSH directory not found"

echo "Authorized keys:"
cat ~/.ssh/authorized_keys 2>/dev/null || echo "No authorized_keys file found"

# Check SSH key permissions
echo -e "\n4. Checking SSH key permissions..."
if [ -f ~/.ssh/id_rsa ]; then
    echo "Private key permissions:"
    ls -la ~/.ssh/id_rsa
    echo "Public key permissions:"
    ls -la ~/.ssh/id_rsa.pub
else
    echo "No SSH keys found in ~/.ssh/"
fi

# Test local SSH connection
echo -e "\n5. Testing local SSH connection..."
if command -v ssh >/dev/null 2>&1; then
    echo "SSH client available"
    # Test connection to localhost
    ssh -o ConnectTimeout=5 -o BatchMode=yes root@localhost echo "Local SSH test successful" 2>/dev/null || echo "Local SSH test failed"
else
    echo "SSH client not available"
fi

echo -e "\n6. Network connectivity test..."
echo "Testing connection to GitHub:"
ping -c 3 github.com 2>/dev/null || echo "Cannot ping GitHub"

echo -e "\n✅ SSH test completed!" 