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

if test -z "$TOKEN_BRIDGE_ADDRESS"
    echo "Missing TOKEN_BRIDGE_ADDRESS"
    exit 1
end

if test -z "$TOKEN_BRIDGE_REGISTRATION_VAAS"
    echo "Missing TOKEN_BRIDGE_REGISTRATION_VAAS"
    exit 1
end

forge script ./forge-scripts/RegisterChainsTokenBridge.s.sol:RegisterChainsTokenBridge \
    --sig "run(address,bytes[])" $TOKEN_BRIDGE_ADDRESS $TOKEN_BRIDGE_REGISTRATION_VAAS \
    --rpc-url $RPC_URL \
    --private-key $MNEMONIC \
    --broadcast
