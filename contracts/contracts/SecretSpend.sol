// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {Verifier} from "./Verifier.sol";

struct ZkProof {
    uint256[8] proof;
    uint256[14] input;
}

contract SecretSpend {
    Verifier internal verifier;
    bytes32 public balancesRoot;

    constructor(address _verifier) {
        verifier = Verifier(_verifier);
    }

    function setBalancesRootForDemo(bytes32 _balancesRoot) external {
        balancesRoot = _balancesRoot;
    }

    function transferPrivately(ZkProof calldata proof) external {
        verifier.verifyProof(proof.proof, proof.input);

        assert(balancesRoot == bytes32(proof.input[0]));
        balancesRoot = bytes32(proof.input[1]);
    }
}
