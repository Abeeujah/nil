// SPDX-License-Identifier: GPL-3.0
//startContract
pragma solidity ^0.8.9;

import "@nilfoundation/smart-contracts/contracts/SmartAccount.sol";
import "@nilfoundation/smart-contracts/contracts/Nil.sol";

/**
 * @title MultiSigSmartAccount
 * @author =nil; Foundation
 * @notice This contract provides a canonical example of a multi-signature smart account on =nil;.
 */
contract MultiSigSmartAccount is NilBase {
    bytes[] pubkeys;

    /**
     * The contract constructor takes an array of public keys.
     * The length of the array cannot exceeed three signatures.
     * @param _publicKeys The array of public keys of the smart account signers.
     */
    constructor(bytes[] memory _publicKeys) payable {
        uint publicKeyLen = _publicKeys.length;
        if (publicKeyLen > 3) {
            revert("MultiSigSmartAccount: too many public keys");
        }
        if (publicKeyLen < 1) {
            revert("MultiSigSmartAccount: too few public keys");
        }
        pubkeys = _publicKeys;
    }

    /**
     * @notice This function parses a signature and defines its r, s, and v components.
     * @param _signatures The signatures from which the function should extract the required signature.
     * @param _pos The starting point from which the analyzed signature starts.
     */
    function parseSignature(
        bytes memory _signatures,
        uint _pos
    ) public pure returns (bytes memory signature) {
        uint offset = _pos * 65;
        bytes32 r;
        bytes32 s;
        uint8 v;
        // The signature format is a compact form of:
        //   {bytes32 r}{bytes32 s}{uint8 v}
        // In this compact form, uint8 is not padded to 32 bytes.
        assembly {
            // solium-disable-line security/no-inline-assembly
            r := mload(add(_signatures, add(32, offset)))
            s := mload(add(_signatures, add(64, offset)))

            // The function loads the last 32 bytes, including 31 bytes
            // of 's'. There is no 'mload8' to do this.
            //
            // 'byte' is not applicable here due to the Solidity parser,
            // so the function uses the second best option, 'and'
            v := and(mload(add(_signatures, add(65, offset))), 0xff)
        }

        if (v < 27) v += 27;

        require(v == 27 || v == 28);

        signature = new bytes(65);

        assembly {
            // solium-disable-line security/no-inline-assembly
            mstore(add(signature, 0x20), r)
            mstore(add(signature, 0x40), s)
            mstore8(add(signature, 0x60), v)
        }

        return signature;
    }

    /**
     * @notice This function acts as a 'wrapper' function for Nil.asyncCallWithTokens().
     * @dev Makes an asynchronous call.
     * @param dst The destination address.
     * @param refundTo The address where to send refund transaction.
     * @param bounceTo The address where to send bounce transaction.
     * @param feeCredit The amount of tokens available to pay all fees during transaction processing.
     * @param tokens Multi-tokens to send.
     * @param value The value to send.
     * @param callData The call data of the called method.
     */
    function asyncCall(
        address dst,
        address refundTo,
        address bounceTo,
        uint feeCredit,
        Nil.Token[] memory tokens,
        uint value,
        bytes calldata callData
    ) public onlyExternal {
        Nil.asyncCallWithTokens(
            dst,
            refundTo,
            bounceTo,
            feeCredit,
            Nil.FORWARD_NONE,
            value,
            tokens,
            callData
        );
    }

    /**
     * @notice This function verifies external signatures and makes it possible to call the smart account externally.
     * @param hash The hash of the external transaction.
     * @param signature The signature of the external transaction.
     */
    function verifyExternal(
        uint256 hash,
        bytes calldata signature
    ) external view returns (bool) {
        uint32 offset = 0;
        uint8 pubkeysLen = uint8(pubkeys.length);
        for (uint i = 0; i < pubkeysLen; i++) {
            bytes memory userSignature = parseSignature(signature, offset);
            offset += 1;
            require(
                Nil.validateSignature(pubkeys[i], hash, userSignature),
                "Invalid signature"
            );
        }
        return true;
    }
}
//endContract
