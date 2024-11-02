#!/usr/bin/env fish

# source the .env file
source .env

# Check for required environment variables
for var in INIT_SIGNERS INIT_CHAIN_ID INIT_GOV_CHAIN_ID INIT_GOV_CONTRACT INIT_EVM_CHAIN_ID MNEMONIC RPC_URL
    if not set -q $var
        echo "Missing $var"
        exit 1
    end
end

# Run the forge script
forge script ./forge-scripts/DeployCore.s.sol:DeployCore \
    --sig "run(address[],uint16,uint16,bytes32,uint256)" $INIT_SIGNERS $INIT_CHAIN_ID $INIT_GOV_CHAIN_ID $INIT_GOV_CONTRACT $INIT_EVM_CHAIN_ID \
    --rpc-url "$RPC_URL" \
    --private-key "$MNEMONIC" \
    --broadcast $FORGE_ARGS

# Read the return info
set returnInfo (cat ./broadcast/DeployCore.s.sol/$INIT_EVM_CHAIN_ID/run-latest.json)

# Extract the address values from 'returnInfo'
set WORMHOLE_ADDRESS (echo $returnInfo | jq -r '.returns.deployedAddress.value')
set SETUP_ADDRESS (echo $returnInfo | jq -r '.returns.setupAddress.value')
set IMPLEMENTATION_ADDRESS (echo $returnInfo | jq -r '.returns.implAddress.value')

# Print the results
echo "-- Wormhole Core Addresses --------------------------------------------------"
echo "| Setup address                | $SETUP_ADDRESS |"
echo "| Implementation address       | $IMPLEMENTATION_ADDRESS |"
echo "| Wormhole address             | $WORMHOLE_ADDRESS |"
echo "-----------------------------------------------------------------------------"
