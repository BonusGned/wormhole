#!/usr/bin/env fish

# MNEMONIC=<redacted> ./sh/registerChainsTokenBridge.fish

source .env

if test -z "$MNEMONIC"
    echo "Missing MNEMONIC"
    exit 1
end

if test -z "$RPC_URL"
    echo "Missing RPC_URL"
    exit 1
end

forge script ./forge-scripts/DeployTestToken.s.sol:DeployTestToken --rpc-url $RPC_URL --private-key $MNEMONIC --broadcast


