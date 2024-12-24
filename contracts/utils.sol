library CUtils {
    function decodeUniV3SwapOutput(bytes memory enc) public pure returns (int256 am0, int256 am1) {
        (am0, am1) = abi.decode(enc, (int256, int256));
    }
}
