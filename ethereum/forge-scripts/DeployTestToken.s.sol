// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.4;
import {ERC20PresetMinterPauser} from "@openzeppelin/contracts/token/ERC20/presets/ERC20PresetMinterPauser.sol";
import {ERC721PresetMinterPauserAutoId} from "@openzeppelin/contracts/token/ERC721/presets/ERC721PresetMinterPauserAutoId.sol";
import {MockWETH9} from "../contracts/bridge/mock/MockWETH9.sol";
import {TokenImplementation} from "../contracts/bridge/token/TokenImplementation.sol";
import "forge-std/Script.sol";

contract DeployTestToken is Script {
    function dryRun() public {
        _deploy();
    }

    function run()
        public
        returns (
            address deployedTokenAddress,
            address deployedNFTaddress,
            address deployedWETHaddress,
            address deployedAccountantTokenAddress
        )
    {
        vm.startBroadcast();
        (
            deployedTokenAddress,
            deployedNFTaddress,
            deployedWETHaddress,
            deployedAccountantTokenAddress
        ) = _deploy();
        vm.stopBroadcast();
    }

    function _deploy()
        internal
        returns (
            address deployedTokenAddress,
            address deployedNFTaddress,
            address deployedWETHaddress,
            address deployedAccountantTokenAddress
        )
    {
        address[] memory accounts = new address[](13);
        accounts[0] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[1] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[2] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[3] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[4] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[5] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[6] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[7] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[8] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[9] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[10] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[11] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        accounts[12] = 0xfBc0163e10FA823E4323B2247da4e7DA7348648B;
        
        ERC20PresetMinterPauser token = new ERC20PresetMinterPauser(
            "Ethereum Test Token",
            "TKN"
        );
        console.log("Token deployed at: ", address(token));

        // mint 1000 units
        token.mint(accounts[0], 1_000_000_000_000_000_000_000);

        ERC721PresetMinterPauserAutoId nft = new ERC721PresetMinterPauserAutoId(
            unicode"Not an APE🐒",
            unicode"APE🐒",
            "https://cloudflare-ipfs.com/ipfs/QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/"
        );

        nft.mint(accounts[0]);
        nft.mint(accounts[0]);

        console.log("NFT deployed at: ", address(nft));

        MockWETH9 mockWeth = new MockWETH9();

        console.log("WETH token deployed at: ", address(mockWeth));

        for(uint16 i=2; i<11; i++) {
            token.mint(accounts[i], 1_000_000_000_000_000_000_000);
        }

        ERC20PresetMinterPauser accountantToken = new ERC20PresetMinterPauser(
            "Accountant Test Token",
            "GA"
        );

        console.log(
            "Accountant test token deployed at: ",
            address(accountantToken)
        );

        // mint 1000 units
        accountantToken.mint(accounts[9], 1_000_000_000_000_000_000_000);

        return (
            address(token),
            address(nft),
            address(mockWeth),
            address(accountantToken)
        );
    }
}
