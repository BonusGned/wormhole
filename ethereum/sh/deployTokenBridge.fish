#!/usr/bin/env fish

# source the .env file
source .env

#Check for required environment variables
for var in INIT_EVM_CHAIN_ID BRIDGE_INIT_CHAIN_ID BRIDGE_INIT_GOV_CHAIN_ID BRIDGE_INIT_GOV_CONTRACT BRIDGE_INIT_WETH BRIDGE_INIT_FINALITY WORMHOLE_ADDRESS MNEMONIC RPC_URL
    if not set -q $var
            echo "Missing $var"
                exit 1
        end
    end

# Run the forge script
forge script ./forge-scripts/DeployTokenBridge.s.sol:DeployTokenBridge \
        --sig "run(uint16,uint16,bytes32,address,uint8,uint256,address)" $BRIDGE_INIT_CHAIN_ID $BRIDGE_INIT_GOV_CHAIN_ID $BRIDGE_INIT_GOV_CONTRACT $BRIDGE_INIT_WETH $BRIDGE_INIT_FINALITY $INIT_EVM_CHAIN_ID $WORMHOLE_ADDRESS \
            --rpc-url "$RPC_URL" \
                --private-key "$MNEMONIC" \
                    --broadcast $FORGE_ARGS

                # Check if the deployment was successful
                if test $status -ne 0
                    echo "Deployment failed"
                        exit 1
                        end

                        # Read the return info
                        set returnInfo (cat ./broadcast/DeployTokenBridge.s.sol/$INIT_EVM_CHAIN_ID/run-latest.json)

                        # Extract the address values from 'returnInfo'
                        set TOKEN_BRIDGE_ADDRESS (echo $returnInfo | jq -r '.returns.deployedAddress.value')
                        set TOKEN_IMPLEMENTATION_ADDRESS (echo $returnInfo | jq -r '.returns.tokenImplementationAddress.value')
                        set TOKEN_BRIDGE_SETUP_ADDRESS (echo $returnInfo | jq -r '.returns.bridgeSetupAddress.value')
                        set TOKEN_BRIDGE_IMPLEMENTATION_ADDRESS (echo $returnInfo | jq -r '.returns.bridgeImplementationAddress.value')

                        # Print the results
                        echo "-- TokenBridge Addresses ----------------------------------------------------"
                        echo "| Token Implementation address | $TOKEN_IMPLEMENTATION_ADDRESS |"
                        echo "| BridgeSetup address          | $TOKEN_BRIDGE_SETUP_ADDRESS |"
                        echo "| BridgeImplementation address | $TOKEN_BRIDGE_IMPLEMENTATION_ADDRESS |"
                        echo "| TokenBridge address          | $TOKEN_BRIDGE_ADDRESS |"
                        echo "-----------------------------------------------------------------------------"

